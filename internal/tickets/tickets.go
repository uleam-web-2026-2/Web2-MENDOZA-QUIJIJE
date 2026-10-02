// Package tickets es el "dominio" de la aplicación: todo lo que tiene que ver
// con un ticket de la mesa de ayuda vive aquí.
//
// El archivo está dividido en tres bloques:
//  1. Modelo:   qué es un Ticket.
//  2. Almacén:  dónde se guardan (en memoria, con candado) y las operaciones
//     internas crear / obtener / listar. Son privadas (minúscula): no saben nada de HTTP.
//  3. Handlers: las funciones que atienden las rutas HTTP (Listar, Obtener, Crear).
//     Son públicas (mayúscula) porque las registra main.go. Ellas validan la
//     entrada, llaman al almacén y responden con el paquete respuesta.
//
// Separar almacén y handlers permite cambiar el almacén por una base de datos
// más adelante sin tocar la parte HTTP.
package tickets

import (
	"encoding/json" // convertir JSON ↔ struct
	"io"            // io.LimitReader: tope de tamaño del cuerpo
	"net/http"
	"strconv" // texto → entero para el {id}
	"strings" // TrimSpace para limpiar el título
	"sync"    // Mutex: el candado del almacén

	"github.com/go-chi/chi/v5" // chi.URLParam para leer el {id} de la ruta

	"github.com/uleam-web-2026-2/web2A-mendoza-quijije/internal/respuesta"
)

// ---------- 1. Modelo ----------

// prioridadesValidas es el conjunto de valores que acepta el campo prioridad.
// Se usa como set: prioridadesValidas["alta"] → true; cualquier otra → false.
var prioridadesValidas = map[string]bool{"baja": true, "media": true, "alta": true}

// Ticket es lo que se guarda y lo que se devuelve al cliente en JSON.
// Las etiquetas `json:"..."` son los nombres de los campos en el JSON (en minúscula).
type Ticket struct {
	ID        int    `json:"id"`        // lo asigna el servidor, autoincremental
	Titulo    string `json:"titulo"`    // mínimo 5 caracteres
	Prioridad string `json:"prioridad"` // baja | media | alta
	Estado    string `json:"estado"`    // siempre "abierto" al crear; el cliente no lo puede fijar
}

// ---------- 2. Almacén en memoria ----------

// Almacen guarda los tickets en un map. Es UNO solo, compartido por todas las
// peticiones (main.go crea una instancia y se la pasa a todas las rutas).
//
// Cada petición HTTP corre en su propia goroutine, así que dos POST simultáneos
// podrían tocar el map al mismo tiempo → datos corruptos o ids repetidos.
// Por eso el mutex: garantiza que solo una goroutine entre a la vez.
type Almacen struct {
	mu     sync.Mutex     // el candado: Lock() antes de tocar ultimo o items, Unlock() al salir
	ultimo int            // último id entregado; el próximo ticket será ultimo+1
	items  map[int]Ticket // los tickets, indexados por id
}

// NuevoAlmacen es el constructor. Siempre crear el almacén con esta función.
func NuevoAlmacen() *Almacen {
	// Un map declarado pero no inicializado (nil) entra en pánico al escribir en él.
	// map[int]Ticket{} lo deja listo para usar.
	return &Almacen{items: map[int]Ticket{}}
}

// crear guarda un ticket nuevo y devuelve la copia con su id asignado.
// Todo lo que toca ultimo o items va entre Lock y Unlock.
func (a *Almacen) crear(titulo, prioridad string) Ticket {
	a.mu.Lock()
	defer a.mu.Unlock() // defer: Unlock garantizado aunque haya return o pánico

	a.ultimo++ // nuevo id (1, 2, 3...)
	t := Ticket{ID: a.ultimo, Titulo: titulo,
		Prioridad: prioridad, Estado: "abierto"} // el estado inicial lo decide el servidor
	a.items[t.ID] = t
	return t
}

// obtener busca un ticket por id. El segundo valor (ok) dice si existía.
func (a *Almacen) obtener(id int) (Ticket, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	t, ok := a.items[id] // leer un map también necesita el candado
	return t, ok
}

// listar devuelve los tickets ordenados por id. El orden de un map en Go
// es aleatorio, por eso se recorre de 1 a ultimo y no el map directamente.
func (a *Almacen) listar() []Ticket {
	a.mu.Lock()
	defer a.mu.Unlock()

	// make(..., 0, n): slice vacío pero NO nil. Importa para el JSON:
	// un slice nil se serializa como null; uno vacío como []. El cliente espera [].
	lista := make([]Ticket, 0, len(a.items))
	for i := 1; i <= a.ultimo; i++ {
		if t, ok := a.items[i]; ok { // el "ok" protege por si algún día se borran tickets
			lista = append(lista, t)
		}
	}
	return lista
}

// ---------- 3. Handlers HTTP (mayúscula: los usa main.go) ----------

// entradaCrear es la forma del JSON que se acepta en POST /tickets.
// Es un struct aparte de Ticket, a propósito: así el cliente NO puede mandar
// "id" ni "estado" (ver DisallowUnknownFields en Crear).
type entradaCrear struct {
	Titulo    string `json:"titulo"`
	Prioridad string `json:"prioridad"`
}

// Listar atiende GET /tickets → 200 con la lista completa (puede ser []).
func (a *Almacen) Listar(w http.ResponseWriter, r *http.Request) {
	respuesta.Exito(w, http.StatusOK, a.listar())
}

// Obtener atiende GET /tickets/{id}.
//
//	200 → el ticket
//	400 id_invalido   → el {id} no es un entero positivo (p. ej. /tickets/abc)
//	404 no_encontrado → es un entero válido pero no existe (p. ej. /tickets/99)
func (a *Almacen) Obtener(w http.ResponseWriter, r *http.Request) {
	// chi.URLParam devuelve el {id} como texto; Atoi lo convierte a entero.
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		respuesta.Error(w, http.StatusBadRequest, "id_invalido",
			"el id debe ser un entero positivo")
		return // importante: si no se retorna, seguiría y escribiría dos respuestas
	}
	t, ok := a.obtener(id)
	if !ok {
		respuesta.Error(w, http.StatusNotFound, "no_encontrado",
			"no existe un ticket con ese id")
		return
	}
	respuesta.Exito(w, http.StatusOK, t)
}

// Crear atiende POST /tickets. Versión de la Fase 3: límite de tamaño + rechazo
// de campos desconocidos.
//
//	201 → ticket creado (con id y estado "abierto")
//	400 cuerpo_malformado → no se pudo ni leer: JSON roto, > 1 MB, campos que no van
//	422 datos_invalidos   → se leyó bien pero viola una regla: título corto, prioridad rara
//
// La diferencia 400 vs 422 es clave: 400 = "no te entendí", 422 = "te entendí
// pero eso no es válido".
func (a *Almacen) Crear(w http.ResponseWriter, r *http.Request) {
	var entrada entradaCrear

	// --- Paso 1: leer el JSON con dos protecciones ---

	// 1 MB de tope (1<<20 = 1.048.576 bytes): sin esto, un cliente podría mandar
	// un título de varios GB y quedarse el servidor sin memoria.
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	// Decisión: rechazar campos que el cliente no debería fijar (p.ej. "estado")
	// en vez de ignorarlos en silencio. Hace visible el intento.
	dec.DisallowUnknownFields()

	if err := dec.Decode(&entrada); err != nil {
		// Cubre: JSON roto, cuerpo que excede 1 MB, y campos desconocidos.
		respuesta.Error(w, http.StatusBadRequest, "cuerpo_malformado",
			"el cuerpo debe ser JSON válido, de hasta 1 MB, solo con los campos titulo y prioridad")
		return
	}

	// --- Paso 2: reglas de negocio → 422 ---
	// El JSON era válido. Ahora se revisa el CONTENIDO: "te entendí,
	// pero lo que me mandaste viola una regla".
	titulo := strings.TrimSpace(entrada.Titulo) // quita espacios al inicio/final: "   abc  " cuenta como "abc"
	if len(titulo) < 5 {
		respuesta.Error(w, http.StatusUnprocessableEntity, "datos_invalidos",
			"el título debe tener al menos 5 caracteres")
		return
	}
	if !prioridadesValidas[entrada.Prioridad] {
		respuesta.Error(w, http.StatusUnprocessableEntity, "datos_invalidos",
			"la prioridad debe ser baja, media o alta")
		return
	}

	// --- Paso 3: todo bien → guardar y responder 201 Created ---
	t := a.crear(titulo, entrada.Prioridad)
	respuesta.Exito(w, http.StatusCreated, t)
}

// Explotar atiende GET /explotar. Existe SOLO para probar el middleware
// Recuperacion: provoca un pánico a propósito y se comprueba que el cliente
// reciba un 500 con envoltura en vez de una conexión cortada.
func Explotar(w http.ResponseWriter, r *http.Request) {
	var t *Ticket // puntero declarado pero sin apuntar a nada → nil
	_ = t.Titulo  // acceder a un campo a través de un puntero nil → pánico
}
