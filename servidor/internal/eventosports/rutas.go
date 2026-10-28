package eventosports

import (
	"eventosports/internal/sesion"
	"github.com/go-chi/chi/v5"
	"net/http"
)

// Rutas es la misma matriz de seguridad en main y en las pruebas.
func (m *Manejador) Rutas(r chi.Router, tokens *sesion.Tokens) {
	r.Get("/eventos", m.listarEventos)
	// La consulta sigue siendo publica. Si se envia token, se valida para
	// incluir la inscripcion del jugador, como en el detalle del hito 1.
	r.With(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "" {
				tokens.Autenticar(next).ServeHTTP(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}).Get("/eventos/{id}", m.verEvento)
	r.Post("/login", m.Sesion.Login)
	r.Group(func(r chi.Router) {
		r.Use(tokens.Autenticar)
		r.With(sesion.SoloRol(sesion.RolOrganizador)).Post("/eventos", m.crearEvento)
		r.With(sesion.SoloRol(sesion.RolJugador)).Post("/eventos/{id}/inscripciones", m.crearInscripcion)
		r.With(sesion.SoloRol(sesion.RolOrganizador)).Get("/eventos/{id}/inscripciones", m.listarInscripciones)
		// Ambos roles pueden entrar; la transicion y la propiedad deciden el permiso.
		r.Patch("/inscripciones/{id}/estado", m.cambiarEstado)
		r.Get("/inscripciones/{id}", m.verInscripcion)
	})
	r.NotFound(func(w http.ResponseWriter, r *http.Request) { responderError(w, 404, "ruta no encontrada") })
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) { responderError(w, 405, "metodo no permitido") })
}
