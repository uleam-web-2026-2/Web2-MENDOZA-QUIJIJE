package negocio

import (
	"eventosports/internal/modelo"
	"strings"
	"time"
)

var deportes = map[string]bool{"futbol": true, "padel": true, "tenis": true, "basquet": true, "natacion": true}
var niveles = map[string]bool{"casual": true, "amateur": true, "competitivo": true}
var estados = map[string]bool{"reservada": true, "pago_por_verificar": true, "confirmada": true, "cancelada": true}

type permiso struct{ Jugador, Organizador bool }

var transiciones = map[string]map[string]permiso{
	"reservada":          {"pago_por_verificar": {Jugador: true}, "confirmada": {Jugador: true}, "cancelada": {Jugador: true, Organizador: true}},
	"pago_por_verificar": {"confirmada": {Organizador: true}, "reservada": {Organizador: true}, "cancelada": {Jugador: true, Organizador: true}},
	"confirmada":         {"cancelada": {Jugador: true, Organizador: true}},
	"cancelada":          {},
}

func TransicionPermitida(de, a string) bool { _, ok := transiciones[de][a]; return ok }

// ValidarCambio es la regla pura que se muestra en la demostracion 5.
// Recibe datos, devuelve una decision; no conoce rutas, peticiones ni bases.
func ValidarCambio(i modelo.Inscripcion, e modelo.Evento, actor Actor, c CambioEstado, confirmadas int64, ahora time.Time) error {
	if !estados[c.Estado] {
		return rechazar(Regla, "estado fuera de la lista")
	}
	jugador := actor.Rol == "jugador" && actor.ID == i.JugadorID
	organizador := actor.Rol == "organizador" && actor.ID == e.OrganizadorID
	if !jugador && !organizador {
		return rechazar(Prohibido, "esta inscripcion no es tuya ni de tus eventos")
	}
	p, ok := transiciones[i.Estado][c.Estado]
	if !ok {
		return rechazar(Regla, "transicion prohibida: de "+i.Estado+" no se pasa a "+c.Estado)
	}
	if (jugador && !p.Jugador) || (organizador && !p.Organizador) {
		return rechazar(Prohibido, "tu rol no puede hacer esta transicion")
	}
	switch {
	case i.Estado == "reservada" && c.Estado == "pago_por_verificar":
		if e.Precio <= 0 {
			return rechazar(Conflicto, "el evento es gratuito: se confirma directamente")
		}
		if strings.TrimSpace(c.Banco) == "" || strings.TrimSpace(c.Referencia) == "" {
			return rechazar(Regla, "faltan banco y referencia")
		}
	case c.Estado == "confirmada":
		if i.Estado == "reservada" && e.Precio > 0 {
			return rechazar(Conflicto, "primero se debe registrar el pago")
		}
		if int64(e.Cupos) <= confirmadas {
			return rechazar(Conflicto, "no quedan cupos")
		}
	case c.Estado == "cancelada":
		if i.Estado == "confirmada" && e.Inicio.Sub(ahora) < 24*time.Hour {
			return rechazar(Conflicto, "faltan menos de 24 horas")
		}
		if i.Estado != "confirmada" && !e.Inicio.After(ahora) {
			return rechazar(Conflicto, "el evento ya empezo")
		}
	}
	return nil
}
