package eventosports

import (
	"eventosports/internal/sesion"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"net/http"
	"strings"
	"time"
)

// Adaptacion de A.1: tambien se validan propiedad, pago, cupos y plazos.
type cuerpoEstado struct {
	Estado     string `json:"estado"`
	Banco      string `json:"banco"`
	Referencia string `json:"referencia"`
}

func (m *Manejador) cambiarEstado(w http.ResponseWriter, r *http.Request) {
	id, ok := idRuta(w, r)
	if !ok {
		return
	}
	var c cuerpoEstado
	if !leerJSON(w, r, &c) {
		return
	}
	if !estados[c.Estado] {
		responderError(w, 422, "estado fuera de la lista")
		return
	}
	s, _ := sesion.De(r)
	var i Inscripcion
	err := m.DB.WithContext(r.Context()).Transaction(func(tx *gorm.DB) error {
		// Crear y confirmar bloquean primero el mismo evento: no se sobrevende el cupo.
		if err := tx.First(&i, id).Error; err != nil {
			return err
		}
		var e Evento
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&e, i.EventoID).Error; err != nil {
			return err
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&i, id).Error; err != nil {
			return err
		}
		jugador := s.Rol == sesion.RolJugador && s.UsuarioID == i.JugadorID
		organizador := s.Rol == sesion.RolOrganizador && s.UsuarioID == e.OrganizadorID
		if !jugador && !organizador {
			return rechazo(403, "esta inscripcion no es tuya ni de tus eventos")
		}
		regla, permitida := transiciones[i.Estado][c.Estado]
		if !permitida {
			return rechazo(422, "transicion prohibida: de "+i.Estado+" no se pasa a "+c.Estado)
		}
		if (jugador && !regla.Jugador) || (organizador && !regla.Organizador) {
			return rechazo(403, "tu rol no puede hacer esta transicion")
		}
		ahora := m.reloj()
		switch {
		case i.Estado == "reservada" && c.Estado == "pago_por_verificar":
			if e.Precio <= 0 {
				return rechazo(409, "el evento es gratuito: se confirma directamente")
			}
			if strings.TrimSpace(c.Banco) == "" || strings.TrimSpace(c.Referencia) == "" {
				return rechazo(422, "faltan banco y referencia")
			}
			p := Pago{InscripcionID: i.ID, Monto: e.Precio, Banco: c.Banco, Referencia: c.Referencia, Creado: ahora}
			if err := tx.Create(&p).Error; err != nil {
				return err
			}
		case c.Estado == "confirmada":
			if i.Estado == "reservada" && e.Precio > 0 {
				return rechazo(409, "primero se debe registrar el pago")
			}
			v, err := vista(tx, e)
			if err != nil {
				return err
			}
			if v.CuposLibres <= 0 {
				return rechazo(409, "no quedan cupos")
			}
			if i.Estado == "pago_por_verificar" {
				var p Pago
				if err := tx.Where("inscripcion_id = ?", i.ID).Order("id DESC").First(&p).Error; err != nil {
					return err
				}
				if err := tx.Model(&p).Update("verificado", true).Error; err != nil {
					return err
				}
			}
		case c.Estado == "cancelada":
			if i.Estado == "confirmada" && e.Inicio.Sub(ahora) < 24*time.Hour {
				return rechazo(409, "faltan menos de 24 horas")
			}
			if i.Estado != "confirmada" && !e.Inicio.After(ahora) {
				return rechazo(409, "el evento ya empezo")
			}
		}
		i.Estado = c.Estado
		return tx.Model(&i).Update("estado", i.Estado).Error
	})
	if !resultado(w, err) {
		responder(w, 200, i)
	}
}
