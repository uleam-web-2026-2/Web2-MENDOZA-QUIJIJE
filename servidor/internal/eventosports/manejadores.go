// Package eventosports adapta HTTP a los casos de uso del negocio.
package eventosports

import (
	"encoding/json"
	"errors"
	"eventosports/internal/negocio"
	"eventosports/internal/sesion"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

type Manejador struct {
	Servicio *negocio.Servicio
	Sesion   *sesion.Manejador
}

func responder(w http.ResponseWriter, codigo int, cuerpo any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(codigo)
	_ = json.NewEncoder(w).Encode(cuerpo)
}
func responderError(w http.ResponseWriter, codigo int, mensaje string) {
	responder(w, codigo, map[string]string{"error": mensaje})
}
func idRuta(w http.ResponseWriter, r *http.Request) (uint, bool) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil || id == 0 {
		responderError(w, 400, "el id debe ser un entero positivo")
		return 0, false
	}
	return uint(id), true
}
func leerJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		responderError(w, 400, "JSON mal formado")
		return false
	}
	return true
}
func actor(r *http.Request) negocio.Actor {
	s, _ := sesion.De(r)
	return negocio.Actor{ID: s.UsuarioID, Rol: s.Rol}
}

// La traduccion a codigos HTTP solo existe en esta capa.
func resultado(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	var e *negocio.Error
	if errors.As(err, &e) {
		codigo := 500
		switch e.Tipo {
		case negocio.Invalido:
			codigo = 400
		case negocio.Prohibido:
			codigo = 403
		case negocio.Conflicto:
			codigo = 409
		case negocio.Regla:
			codigo = 422
		}
		responderError(w, codigo, e.Mensaje)
	} else if errors.Is(err, negocio.ErrNoExiste) {
		responderError(w, 404, "el recurso no existe")
	} else {
		responderError(w, 500, "no se pudo acceder a los datos")
	}
	return true
}
func (m *Manejador) listarEventos(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	v, err := m.Servicio.ListarEventos(r.Context(), negocio.Filtro{Deporte: q.Get("deporte"), Nivel: q.Get("nivel"), Ciudad: q.Get("ciudad")})
	if !resultado(w, err) {
		responder(w, 200, v)
	}
}
func (m *Manejador) verEvento(w http.ResponseWriter, r *http.Request) {
	id, ok := idRuta(w, r)
	if !ok {
		return
	}
	v, err := m.Servicio.VerEvento(r.Context(), id, actor(r))
	if !resultado(w, err) {
		responder(w, 200, v)
	}
}
func (m *Manejador) crearEvento(w http.ResponseWriter, r *http.Request) {
	var c negocio.CrearEvento
	if !leerJSON(w, r, &c) {
		return
	}
	v, err := m.Servicio.CrearEvento(r.Context(), c, actor(r))
	if !resultado(w, err) {
		responder(w, 201, v)
	}
}
func (m *Manejador) crearInscripcion(w http.ResponseWriter, r *http.Request) {
	id, ok := idRuta(w, r)
	if !ok {
		return
	}
	v, err := m.Servicio.CrearInscripcion(r.Context(), id, actor(r))
	if !resultado(w, err) {
		responder(w, 201, v)
	}
}
func (m *Manejador) verInscripcion(w http.ResponseWriter, r *http.Request) {
	id, ok := idRuta(w, r)
	if !ok {
		return
	}
	v, err := m.Servicio.VerInscripcion(r.Context(), id, actor(r))
	if !resultado(w, err) {
		responder(w, 200, v)
	}
}
func (m *Manejador) listarInscripciones(w http.ResponseWriter, r *http.Request) {
	id, ok := idRuta(w, r)
	if !ok {
		return
	}
	v, err := m.Servicio.ListarInscripciones(r.Context(), id, actor(r))
	if !resultado(w, err) {
		responder(w, 200, v)
	}
}
