package eventosports

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Servidor guarda los datos en memoria y atiende los seis endpoints de la sección 7 de la ficha.
type Servidor struct {
	mu             sync.Mutex
	usuarios       map[uint]Usuario
	sedes          map[uint]Sede
	eventos        map[uint]Evento
	inscripciones  map[uint]Inscripcion
	pagos          []Pago
	sigEvento      uint
	sigInscripcion uint
	sigPago        uint
	ahora          func() time.Time // se puede reemplazar en las pruebas
}

// NuevoServidor carga unos datos de ejemplo: un organizador (7), un jugador (12), una sede (3) y dos eventos (25 y 26).
func NuevoServidor() *Servidor {
	zona := time.FixedZone("ECT", -5*3600)
	return &Servidor{
		usuarios: map[uint]Usuario{
			7:  {ID: 7, Nombre: "Club Deportivo Manta Norte", Correo: "manta.norte@correo.com", Rol: "organizador"},
			12: {ID: 12, Nombre: "Ana Zambrano", Correo: "ana.zambrano@correo.com", Rol: "jugador"},
		},
		sedes: map[uint]Sede{
			3: {ID: 3, Nombre: "Cancha sintética El Palmar", Ciudad: "Manta", Referencia: "Junto al parque, frente a la tienda de la esquina"},
		},
		eventos: map[uint]Evento{
			25: {ID: 25, Titulo: "Fútbol 7 mixto, jueves noche", Deporte: "futbol", Nivel: "casual", Inicio: time.Date(2026, 10, 15, 20, 0, 0, 0, zona), Cupos: 14, Precio: 2.50, OrganizadorID: 7, SedeID: 3},
			26: {ID: 26, Titulo: "Pádel dobles, sábado en la mañana", Deporte: "padel", Nivel: "amateur", Inicio: time.Date(2026, 10, 17, 9, 0, 0, 0, zona), Cupos: 8, Precio: 0, OrganizadorID: 7, SedeID: 3},
		},
		inscripciones:  map[uint]Inscripcion{},
		sigEvento:      27,
		sigInscripcion: 118,
		sigPago:        60,
		ahora:          time.Now,
	}
}

// ---------- utilidades

func responder(w http.ResponseWriter, codigo int, cuerpo any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(codigo)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(cuerpo)
}

func responderError(w http.ResponseWriter, codigo int, mensaje string) {
	responder(w, codigo, map[string]string{"error": mensaje})
}

// identidad lee X-Usuario-ID y X-Rol (el inicio de sesión real es de un hito posterior).
func identidad(r *http.Request) (uint, string) {
	id, _ := strconv.ParseUint(r.Header.Get("X-Usuario-ID"), 10, 64)
	rol := r.Header.Get("X-Rol")
	if rol != "jugador" && rol != "organizador" {
		return 0, ""
	}
	return uint(id), rol
}

// idDeLaRuta convierte el {id} de la ruta; si no es un número, el manejador responde 400.
func idDeLaRuta(texto string) (uint, bool) {
	n, err := strconv.ParseUint(texto, 10, 64)
	return uint(n), err == nil
}

type eventoVista struct {
	Evento
	Sede          Sede         `json:"sede"`
	CuposLibres   int          `json:"cupos_libres"`
	MiInscripcion *Inscripcion `json:"mi_inscripcion,omitempty"`
}

// cuposLibres cuenta solo las inscripciones confirmadas (llamar con el candado tomado).
func (s *Servidor) cuposLibres(e Evento) int {
	usados := 0
	for _, i := range s.inscripciones {
		if i.EventoID == e.ID && i.Estado == "confirmada" {
			usados++
		}
	}
	return e.Cupos - usados
}

func (s *Servidor) vista(e Evento) eventoVista {
	return eventoVista{Evento: e, Sede: s.sedes[e.SedeID], CuposLibres: s.cuposLibres(e)}
}

// ---------- GET /eventos

func (s *Servidor) listarEventos(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	deporte, nivel, ciudad := q.Get("deporte"), q.Get("nivel"), q.Get("ciudad")
	if deporte != "" && !deportes[deporte] {
		responderError(w, http.StatusUnprocessableEntity, "deporte fuera de la lista: futbol, padel, tenis, basquet, natacion")
		return
	}
	if nivel != "" && !niveles[nivel] {
		responderError(w, http.StatusUnprocessableEntity, "nivel fuera de la lista: casual, amateur, competitivo")
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	lista := []eventoVista{}
	for _, e := range s.eventos {
		if deporte != "" && e.Deporte != deporte {
			continue
		}
		if nivel != "" && e.Nivel != nivel {
			continue
		}
		if ciudad != "" && !strings.EqualFold(s.sedes[e.SedeID].Ciudad, ciudad) {
			continue
		}
		lista = append(lista, s.vista(e))
	}
	sort.Slice(lista, func(a, b int) bool { return lista[a].ID < lista[b].ID })
	responder(w, http.StatusOK, lista)
}

// ---------- POST /eventos

type cuerpoEvento struct {
	Titulo  string    `json:"titulo"`
	Deporte string    `json:"deporte"`
	Nivel   string    `json:"nivel"`
	Inicio  time.Time `json:"inicio"`
	Cupos   int       `json:"cupos"`
	Precio  float64   `json:"precio"`
	SedeID  uint      `json:"sede_id"`
}

func (s *Servidor) crearEvento(w http.ResponseWriter, r *http.Request) {
	uid, rol := identidad(r)
	if rol != "organizador" {
		responderError(w, http.StatusForbidden, "solo un organizador puede publicar eventos")
		return
	}
	var c cuerpoEvento
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		responderError(w, http.StatusBadRequest, "JSON mal formado")
		return
	}
	// Todas las reglas se validan antes de tocar los datos.
	mensaje := ""
	switch {
	case strings.TrimSpace(c.Titulo) == "":
		mensaje = "titulo no puede estar vacío"
	case !deportes[c.Deporte]:
		mensaje = "deporte fuera de la lista: futbol, padel, tenis, basquet, natacion"
	case !niveles[c.Nivel]:
		mensaje = "nivel fuera de la lista: casual, amateur, competitivo"
	case c.Cupos <= 0:
		mensaje = "cupos debe ser mayor que 0"
	case c.Precio < 0:
		mensaje = "precio no puede ser negativo"
	case !c.Inicio.After(s.ahora()):
		mensaje = "inicio debe estar en el futuro"
	}
	if mensaje != "" {
		responderError(w, http.StatusUnprocessableEntity, mensaje)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.sedes[c.SedeID]; !ok {
		responderError(w, http.StatusUnprocessableEntity, "la sede no existe")
		return
	}
	e := Evento{ID: s.sigEvento, Titulo: c.Titulo, Deporte: c.Deporte, Nivel: c.Nivel, Inicio: c.Inicio,
		Cupos: c.Cupos, Precio: c.Precio, OrganizadorID: uid, SedeID: c.SedeID}
	s.sigEvento++
	s.eventos[e.ID] = e
	responder(w, http.StatusCreated, s.vista(e))
}

// ---------- GET /eventos/{id}

func (s *Servidor) verEvento(w http.ResponseWriter, r *http.Request, idTexto string) {
	id, ok := idDeLaRuta(idTexto)
	if !ok {
		responderError(w, http.StatusBadRequest, "el id del evento debe ser un número")
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	e, existe := s.eventos[id]
	if !existe {
		responderError(w, http.StatusNotFound, "el evento no existe")
		return
	}
	v := s.vista(e)
	if uid, rol := identidad(r); rol == "jugador" {
		var ultima *Inscripcion
		for _, i := range s.inscripciones {
			if i.EventoID == e.ID && i.JugadorID == uid && (ultima == nil || i.ID > ultima.ID) {
				copia := i
				ultima = &copia
			}
		}
		v.MiInscripcion = ultima
	}
	responder(w, http.StatusOK, v)
}

// ---------- POST /eventos/{id}/inscripciones

func (s *Servidor) crearInscripcion(w http.ResponseWriter, r *http.Request, idTexto string) {
	uid, rol := identidad(r)
	if rol != "jugador" {
		responderError(w, http.StatusForbidden, "solo un jugador puede inscribirse")
		return
	}
	id, ok := idDeLaRuta(idTexto)
	if !ok {
		responderError(w, http.StatusBadRequest, "el id del evento debe ser un número")
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	e, existe := s.eventos[id]
	if !existe {
		responderError(w, http.StatusNotFound, "el evento no existe")
		return
	}
	if !e.Inicio.After(s.ahora()) {
		responderError(w, http.StatusConflict, "el evento ya empezó")
		return
	}
	for _, i := range s.inscripciones {
		if i.EventoID == e.ID && i.JugadorID == uid && i.Estado != "cancelada" {
			responderError(w, http.StatusConflict, "ya tienes una inscripción activa en este evento")
			return
		}
	}
	if s.cuposLibres(e) <= 0 {
		responderError(w, http.StatusConflict, "no quedan cupos")
		return
	}
	i := Inscripcion{ID: s.sigInscripcion, EventoID: e.ID, JugadorID: uid, Estado: "reservada", Creado: s.ahora()}
	s.sigInscripcion++
	s.inscripciones[i.ID] = i
	responder(w, http.StatusCreated, i)
}

// ---------- PATCH /inscripciones/{id}/estado

type cuerpoEstado struct {
	Estado     string `json:"estado"`
	Banco      string `json:"banco"`
	Referencia string `json:"referencia"`
}

func (s *Servidor) cambiarEstado(w http.ResponseWriter, r *http.Request, idTexto string) {
	uid, rol := identidad(r)
	if rol == "" {
		responderError(w, http.StatusForbidden, "falta identificarse con X-Usuario-ID y X-Rol")
		return
	}
	id, ok := idDeLaRuta(idTexto)
	if !ok {
		responderError(w, http.StatusBadRequest, "el id de la inscripción debe ser un número")
		return
	}
	var c cuerpoEstado
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		responderError(w, http.StatusBadRequest, "JSON mal formado")
		return
	}
	if !estados[c.Estado] {
		responderError(w, http.StatusUnprocessableEntity, "estado fuera de la lista: reservada, pago_por_verificar, confirmada, cancelada")
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	i, existe := s.inscripciones[id]
	if !existe {
		responderError(w, http.StatusNotFound, "la inscripción no existe")
		return
	}
	e := s.eventos[i.EventoID]
	esJugador := rol == "jugador" && uid == i.JugadorID
	esOrganizador := rol == "organizador" && uid == e.OrganizadorID
	if !esJugador && !esOrganizador {
		responderError(w, http.StatusForbidden, "esta inscripción no es tuya ni de uno de tus eventos")
		return
	}
	regla, permitida := transiciones[i.Estado][c.Estado]
	if !permitida {
		responderError(w, http.StatusConflict, fmt.Sprintf("transición prohibida: de %s no se pasa a %s", i.Estado, c.Estado))
		return
	}
	if (esJugador && !regla.Jugador) || (esOrganizador && !regla.Organizador) {
		responderError(w, http.StatusForbidden, "tu rol no puede hacer esta transición")
		return
	}
	ahora := s.ahora()
	switch {
	case i.Estado == "reservada" && c.Estado == "pago_por_verificar":
		if e.Precio <= 0 {
			responderError(w, http.StatusConflict, "el evento es gratuito: se confirma directamente")
			return
		}
		if strings.TrimSpace(c.Banco) == "" || strings.TrimSpace(c.Referencia) == "" {
			responderError(w, http.StatusUnprocessableEntity, "faltan banco y referencia de la transferencia")
			return
		}
		s.pagos = append(s.pagos, Pago{ID: s.sigPago, InscripcionID: i.ID, Monto: e.Precio, Banco: c.Banco, Referencia: c.Referencia, Creado: ahora})
		s.sigPago++
	case i.Estado == "reservada" && c.Estado == "confirmada":
		if e.Precio > 0 {
			responderError(w, http.StatusConflict, "el evento es de pago: primero se manda banco y referencia")
			return
		}
		if s.cuposLibres(e) <= 0 {
			responderError(w, http.StatusConflict, "no quedan cupos")
			return
		}
	case i.Estado == "pago_por_verificar" && c.Estado == "confirmada":
		if s.cuposLibres(e) <= 0 {
			responderError(w, http.StatusConflict, "no quedan cupos")
			return
		}
		for k := len(s.pagos) - 1; k >= 0; k-- { // marca el último pago de la inscripción como verificado
			if s.pagos[k].InscripcionID == i.ID {
				s.pagos[k].Verificado = true
				break
			}
		}
	case c.Estado == "cancelada":
		if i.Estado == "confirmada" && e.Inicio.Sub(ahora) < 24*time.Hour {
			responderError(w, http.StatusConflict, "faltan menos de 24 horas para el evento")
			return
		}
		if i.Estado != "confirmada" && !e.Inicio.After(ahora) {
			responderError(w, http.StatusConflict, "el evento ya empezó")
			return
		}
	}
	i.Estado = c.Estado
	s.inscripciones[i.ID] = i
	responder(w, http.StatusOK, i)
}

// ---------- GET /eventos/{id}/inscripciones

type jugadorVista struct {
	ID     uint   `json:"id"`
	Nombre string `json:"nombre,omitempty"`
}

type inscripcionVista struct {
	Inscripcion
	Jugador    jugadorVista `json:"jugador"`
	UltimoPago *Pago        `json:"ultimo_pago,omitempty"`
}

func (s *Servidor) listarInscripciones(w http.ResponseWriter, r *http.Request, idTexto string) {
	uid, rol := identidad(r)
	if rol != "organizador" {
		responderError(w, http.StatusForbidden, "solo el organizador del evento puede ver sus inscripciones")
		return
	}
	id, ok := idDeLaRuta(idTexto)
	if !ok {
		responderError(w, http.StatusBadRequest, "el id del evento debe ser un número")
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	e, existe := s.eventos[id]
	if !existe {
		responderError(w, http.StatusNotFound, "el evento no existe")
		return
	}
	if e.OrganizadorID != uid {
		responderError(w, http.StatusForbidden, "este evento es de otro organizador")
		return
	}
	lista := []inscripcionVista{}
	for _, i := range s.inscripciones {
		if i.EventoID != e.ID {
			continue
		}
		v := inscripcionVista{Inscripcion: i, Jugador: jugadorVista{ID: i.JugadorID, Nombre: s.usuarios[i.JugadorID].Nombre}}
		for k := len(s.pagos) - 1; k >= 0; k-- {
			if s.pagos[k].InscripcionID == i.ID {
				copia := s.pagos[k]
				v.UltimoPago = &copia
				break
			}
		}
		lista = append(lista, v)
	}
	sort.Slice(lista, func(a, b int) bool { return lista[a].ID < lista[b].ID })
	responder(w, http.StatusOK, lista)
}
