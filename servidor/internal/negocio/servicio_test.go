package negocio

import (
	"context"
	"errors"
	"testing"
	"time"

	"eventosports/internal/modelo"
)

type almacenFalso struct {
	Almacen
	operacion *operacionFalsa
}

func (a *almacenFalso) Transaccion(_ context.Context, ejecutar func(Operacion) error) error {
	return ejecutar(a.operacion)
}

type operacionFalsa struct {
	Operacion
	pagoCreado bool
}

func (o *operacionFalsa) Inscripcion(uint) (modelo.Inscripcion, error) {
	return modelo.Inscripcion{ID: 1, EventoID: 25, JugadorID: 1, Estado: "reservada"}, nil
}

func (o *operacionFalsa) EventoBloqueado(uint) (modelo.Evento, error) {
	return modelo.Evento{ID: 25, OrganizadorID: 2, Precio: 2.50, Cupos: 10, Inicio: time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)}, nil
}

func (o *operacionFalsa) InscripcionBloqueada(id uint) (modelo.Inscripcion, error) {
	return o.Inscripcion(id)
}

func (o *operacionFalsa) Confirmadas(uint) (int64, error) { return 0, nil }

func (o *operacionFalsa) CrearPago(*modelo.Pago) error {
	o.pagoCreado = true
	return nil
}

var errEstadoForzado = errors.New("fallo forzado al guardar el estado")

func (o *operacionFalsa) GuardarEstado(modelo.Inscripcion) error {
	if !o.pagoCreado {
		return errors.New("el servicio no creó el pago antes del cambio de estado")
	}
	return errEstadoForzado
}

func TestServicioPropagaFalloConAlmacenFalso(t *testing.T) {
	operacion := &operacionFalsa{}
	servicio := &Servicio{
		Almacen: &almacenFalso{operacion: operacion},
		Ahora:   func() time.Time { return time.Date(2026, 11, 5, 0, 0, 0, 0, time.UTC) },
	}
	_, err := servicio.CambiarEstado(context.Background(), 1,
		CambioEstado{Estado: "pago_por_verificar", Banco: "Banco", Referencia: "PRUEBA"},
		Actor{ID: 1, Rol: "jugador"})
	if !errors.Is(err, errEstadoForzado) {
		t.Fatalf("el servicio debe devolver el fallo del repositorio: %v", err)
	}
}
