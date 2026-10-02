// main.go — punto de entrada del servidor de la mesa de ayuda.
//
// Aquí NO hay lógica de negocio. Este archivo solo hace tres cosas:
//  1. Crea el router (chi) y le cuelga los middlewares.
//  2. Define qué ruta atiende qué handler.
//  3. Levanta el servidor en el puerto :8080.
//
// El resto del código vive en internal/:
//   - internal/respuesta  → formato único de las respuestas JSON
//   - internal/middleware → Registro (log) y Recuperacion (atrapa pánicos)
//   - internal/tickets    → el modelo Ticket, el almacén en memoria y los handlers
package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5" // router HTTP: rutas con parámetros, middlewares, 404/405 personalizables

	// Paquetes propios. La ruta larga es el nombre del módulo (1.ª línea de go.mod)
	// seguido de la carpeta; así es como Go localiza paquetes internos.
	"github.com/uleam-web-2026-2/web2A-mendoza-quijije/internal/middleware"
	"github.com/uleam-web-2026-2/web2A-mendoza-quijije/internal/respuesta"
	"github.com/uleam-web-2026-2/web2A-mendoza-quijije/internal/tickets"
)

func main() {
	// ---------- 1. Router ----------
	r := chi.NewRouter()

	// ---------- 2. Middlewares ----------
	// Un middleware envuelve a los handlers: corre código antes y/o después de ellos.
	// r.Use los aplica a TODAS las rutas. El orden importa:
	//
	// Registro AFUERA, Recuperación ADENTRO: la cadena queda Registro(Recuperacion(handler)).
	// Así, si un handler entra en pánico, Recuperación lo atrapa y responde 500,
	// y Registro (que sigue vivo por fuera) SÍ alcanza a loguear esa petición con su 500.
	// Al revés, el pánico salta por encima de Registro (que no tiene recover) y esa
	// petición nunca queda en el log aunque el cliente sí reciba el 500.
	r.Use(middleware.Registro)
	r.Use(middleware.Recuperacion)

	// ---------- 3. Respuestas de error del propio router ----------
	// Chi responde 404 y 405 en texto plano por defecto. Se sobrescriben para que
	// TODO lo que salga del servidor tenga la misma envoltura JSON {ok, datos/error}.

	// 404: la ruta no existe (p. ej. GET /usuarios).
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		respuesta.Error(w, http.StatusNotFound, "ruta_inexistente",
			"la ruta no existe")
	})
	// 405: la ruta existe pero no con ese método (p. ej. DELETE /tickets/1).
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		respuesta.Error(w, http.StatusMethodNotAllowed, "metodo_no_permitido",
			"el método no está permitido en esta ruta")
	})

	// ---------- 4. Rutas ----------

	// GET / — página de bienvenida de la API. Sin esta ruta, abrir
	// localhost:8080 en el navegador daba 404. Devuelve qué rutas existen.
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		respuesta.Exito(w, http.StatusOK, map[string]any{
			"servicio": "mesa de ayuda",
			"rutas":    []string{"GET /tickets", "POST /tickets", "GET /tickets/{id}"},
		})
	})

	// Un solo almacén compartido por todas las peticiones (por eso lleva mutex adentro).
	almacen := tickets.NuevoAlmacen()

	r.Get("/tickets", almacen.Listar)       // GET  /tickets      → lista todos (200)
	r.Post("/tickets", almacen.Crear)       // POST /tickets      → crea uno (201 / 400 / 422)
	r.Get("/tickets/{id}", almacen.Obtener) // GET  /tickets/{id} → uno por id (200 / 400 / 404). {id} es parámetro de ruta.
	r.Get("/explotar", tickets.Explotar)    // GET  /explotar     → provoca un pánico a propósito; solo para probar Recuperacion (500)

	// ---------- 5. Arranque ----------
	log.Println("mesa de ayuda en :8080")
	// ListenAndServe bloquea para siempre atendiendo peticiones. Solo retorna si
	// falla (p. ej. puerto ocupado); en ese caso log.Fatal imprime el error y sale.
	log.Fatal(http.ListenAndServe(":8080", r))
}
