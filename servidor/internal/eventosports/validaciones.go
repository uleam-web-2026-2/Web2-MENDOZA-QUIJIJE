package eventosports

// Listas cerradas (campos «uno de» de la sección 4 de la ficha). Se validan antes de tocar los datos.
var deportes = map[string]bool{"futbol": true, "padel": true, "tenis": true, "basquet": true, "natacion": true}
var niveles = map[string]bool{"casual": true, "amateur": true, "competitivo": true}
var estados = map[string]bool{"reservada": true, "pago_por_verificar": true, "confirmada": true, "cancelada": true}

// quien puede disparar cada transición (sección 5 de la ficha).
type quien struct{ Jugador, Organizador bool }

// transiciones[de][a]. Lo que no está aquí está prohibido; de «cancelada» no sale nada (estado final).
var transiciones = map[string]map[string]quien{
	"reservada": {
		"pago_por_verificar": {Jugador: true},
		"confirmada":         {Jugador: true},
		"cancelada":          {Jugador: true, Organizador: true},
	},
	"pago_por_verificar": {
		"confirmada": {Organizador: true},
		"reservada":  {Organizador: true},
		"cancelada":  {Jugador: true, Organizador: true},
	},
	"confirmada": {
		"cancelada": {Jugador: true, Organizador: true},
	},
}

// transicionPermitida consulta la tabla de transiciones: es lo que bloquea cancelada → confirmada.
func transicionPermitida(de, a string) bool {
	_, ok := transiciones[de][a]
	return ok
}
