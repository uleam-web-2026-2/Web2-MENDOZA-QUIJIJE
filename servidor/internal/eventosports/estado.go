package eventosports

import (
	"eventosports/internal/negocio"
	"net/http"
)

// cambiarEstado solo entiende HTTP: no consulta GORM ni decide transiciones.
func (m *Manejador) cambiarEstado(w http.ResponseWriter, r *http.Request) {
	id, ok := idRuta(w, r)
	if !ok {
		return
	}
	var c negocio.CambioEstado
	if !leerJSON(w, r, &c) {
		return
	}
	i, err := m.Servicio.CambiarEstado(r.Context(), id, c, actor(r))
	if !resultado(w, err) {
		responder(w, 200, i)
	}
}
