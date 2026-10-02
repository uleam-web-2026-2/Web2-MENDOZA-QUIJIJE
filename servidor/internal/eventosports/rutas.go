package eventosports

import (
	"net/http"
	"strings"
)

// Rutas registra los seis endpoints de la sección 7 de la ficha (enrutado a mano, funciona con cualquier versión de Go).
func (s *Servidor) Rutas() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		switch {
		case len(p) == 1 && p[0] == "eventos":
			switch r.Method {
			case http.MethodGet:
				s.listarEventos(w, r)
			case http.MethodPost:
				s.crearEvento(w, r)
			default:
				responderError(w, http.StatusMethodNotAllowed, "método no permitido")
			}
		case len(p) == 2 && p[0] == "eventos" && r.Method == http.MethodGet:
			s.verEvento(w, r, p[1])
		case len(p) == 3 && p[0] == "eventos" && p[2] == "inscripciones" && r.Method == http.MethodPost:
			s.crearInscripcion(w, r, p[1])
		case len(p) == 3 && p[0] == "eventos" && p[2] == "inscripciones" && r.Method == http.MethodGet:
			s.listarInscripciones(w, r, p[1])
		case len(p) == 3 && p[0] == "inscripciones" && p[2] == "estado" && r.Method == http.MethodPatch:
			s.cambiarEstado(w, r, p[1])
		default:
			responderError(w, http.StatusNotFound, "ruta no encontrada")
		}
	})
}
