package negocio

import (
	"errors"
	"eventosports/internal/modelo"
	"testing"
	"time"
)

// Pruebas sin servidor ni base: la regla se puede explicar y comprobar sola.
func TestValidarCambioSinHTTPNiBase(t *testing.T) {
	ahora := time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)
	for _, c := range []struct {
		nombre, de, a, rol string
		precio             float64
		cupos              int
		confirmadas        int64
		anticipacion       time.Duration
		tipo               TipoError
	}{
		{"cancelada es final", "cancelada", "confirmada", "organizador", 2.5, 14, 0, 48 * time.Hour, Regla},
		{"jugador no verifica", "pago_por_verificar", "confirmada", "jugador", 2.5, 14, 0, 48 * time.Hour, Prohibido},
		{"organizador verifica", "pago_por_verificar", "confirmada", "organizador", 2.5, 14, 0, 48 * time.Hour, ""},
		{"gratis se confirma", "reservada", "confirmada", "jugador", 0, 14, 0, 48 * time.Hour, ""},
		{"pagado exige comprobante", "reservada", "confirmada", "jugador", 2.5, 14, 0, 48 * time.Hour, Conflicto},
		{"no sobrevende", "pago_por_verificar", "confirmada", "organizador", 2.5, 1, 1, 48 * time.Hour, Conflicto},
		{"cancelacion exactamente 24h", "confirmada", "cancelada", "jugador", 0, 14, 0, 24 * time.Hour, ""},
		{"cancelacion fuera de plazo", "confirmada", "cancelada", "jugador", 0, 14, 0, 23 * time.Hour, Conflicto},
	} {
		t.Run(c.nombre, func(t *testing.T) {
			a := Actor{ID: 1, Rol: c.rol}
			if c.rol == "organizador" {
				a.ID = 2
			}
			i := modelo.Inscripcion{JugadorID: 1, Estado: c.de}
			e := modelo.Evento{OrganizadorID: 2, Precio: c.precio, Cupos: c.cupos, Inicio: ahora.Add(c.anticipacion)}
			err := ValidarCambio(i, e, a, CambioEstado{Estado: c.a}, c.confirmadas, ahora)
			if c.tipo == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			var rechazo *Error
			if !errors.As(err, &rechazo) || rechazo.Tipo != c.tipo {
				t.Fatalf("esperado %s; recibido %v", c.tipo, err)
			}
		})
	}
}
