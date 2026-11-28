package negocio

import (
	"context"
	"errors"
	"eventosports/internal/modelo"
	"testing"
	"time"
)

// avisadorFalso anota a quién se avisó, sin WebSocket.
type avisadorFalso struct {
	quien  []uint
	avisos []Aviso
}

func (a *avisadorFalso) Enviar(usuarioID uint, av Aviso) {
	a.quien = append(a.quien, usuarioID)
	a.avisos = append(a.avisos, av)
}

// almacenFalso cumple Almacen sin base de datos: Transaccion corre la
// función con una operacionFalsa que guarda una inscripción y su evento en
// memoria. Los métodos que no se escriben aquí no se usan en estas pruebas.
type almacenFalso struct {
	Almacen
	op *operacionFalsa
}

func (a *almacenFalso) Transaccion(_ context.Context, f func(Operacion) error) error {
	return f(a.op)
}

type operacionFalsa struct {
	Operacion
	ins        modelo.Inscripcion
	ev         modelo.Evento
	pago       modelo.Pago
	errGuardar error // si no es nil, GuardarEstado falla (la base se cayó)
}

func (o *operacionFalsa) Inscripcion(uint) (modelo.Inscripcion, error) { return o.ins, nil }
func (o *operacionFalsa) InscripcionBloqueada(uint) (modelo.Inscripcion, error) {
	return o.ins, nil
}
func (o *operacionFalsa) EventoBloqueado(uint) (modelo.Evento, error) { return o.ev, nil }
func (o *operacionFalsa) Confirmadas(uint) (int64, error)             { return 0, nil }
func (o *operacionFalsa) UltimoPago(uint) (modelo.Pago, error)        { return o.pago, nil }
func (o *operacionFalsa) VerificarPago(uint) error                    { return nil }
func (o *operacionFalsa) GuardarEstado(i modelo.Inscripcion) error {
	if o.errGuardar != nil {
		return o.errGuardar
	}
	o.ins = i
	return nil
}

// El caso del paso 3.2 en EventoSports: ana (jugadora, id 1) tiene la
// inscripción 7 en un evento pagado de luis (organizador, id 2).
var (
	organizadorLuis = Actor{ID: 2, Rol: "organizador"}
	jugadoraEva     = Actor{ID: 3, Rol: "jugador"} // del mismo rol que ana
)

func servicioConAvisos(estado string) (*Servicio, *operacionFalsa, *avisadorFalso) {
	ahora := time.Date(2026, 11, 27, 20, 0, 0, 0, time.UTC)
	op := &operacionFalsa{
		ins:  modelo.Inscripcion{ID: 7, EventoID: 25, JugadorID: 1, Estado: estado},
		ev:   modelo.Evento{ID: 25, OrganizadorID: 2, Precio: 2.5, Cupos: 14, Inicio: ahora.Add(30 * 24 * time.Hour)},
		pago: modelo.Pago{ID: 3, InscripcionID: 7},
	}
	av := &avisadorFalso{}
	s := &Servicio{Almacen: &almacenFalso{op: op}, Ahora: func() time.Time { return ahora }}
	s.AvisarCon(av)
	return s, op, av
}

func TestCambiarEstadoAvisaAlJugadorDueno(t *testing.T) {
	s, op, av := servicioConAvisos("pago_por_verificar")
	if _, err := s.CambiarEstado(context.Background(), 7, CambioEstado{Estado: "confirmada"}, organizadorLuis); err != nil {
		t.Fatal(err)
	}
	if op.ins.Estado != "confirmada" {
		t.Fatalf("no se guardó el estado: %s", op.ins.Estado)
	}
	if len(av.quien) != 1 || av.quien[0] != 1 {
		t.Fatalf("se esperaba un aviso para ana (id 1) y se avisó a %v", av.quien)
	}
	if av.avisos[0] != (Aviso{Tipo: "estado_cambiado", ID: 7, Estado: "confirmada"}) {
		t.Fatalf("el aviso tiene que traer tipo, id y estado: %+v", av.avisos[0])
	}
}

func TestTransicionProhibidaNoAvisa(t *testing.T) {
	s, _, av := servicioConAvisos("cancelada")
	if _, err := s.CambiarEstado(context.Background(), 7, CambioEstado{Estado: "confirmada"}, organizadorLuis); err == nil {
		t.Fatal("de cancelada a confirmada tenía que fallar")
	}
	if len(av.quien) != 0 {
		t.Fatalf("no cambió nada y se avisó a %v", av.quien)
	}
}

func TestSiGuardarFallaNoAvisa(t *testing.T) {
	s, op, av := servicioConAvisos("pago_por_verificar")
	op.errGuardar = errors.New("se cayó la base")
	if _, err := s.CambiarEstado(context.Background(), 7, CambioEstado{Estado: "confirmada"}, organizadorLuis); err == nil {
		t.Fatal("si GuardarEstado falla, CambiarEstado tiene que fallar")
	}
	if len(av.quien) != 0 {
		t.Fatalf("no quedó en la base y se avisó a %v", av.quien)
	}
}

func TestJugadorAjenoNoCambiaNiAvisa(t *testing.T) {
	s, _, av := servicioConAvisos("reservada")
	_, err := s.CambiarEstado(context.Background(), 7, CambioEstado{Estado: "cancelada"}, jugadoraEva)
	var rechazo *Error
	if !errors.As(err, &rechazo) || rechazo.Tipo != Prohibido {
		t.Fatalf("eva no puede cancelar la inscripción de ana: %v", err)
	}
	if len(av.quien) != 0 {
		t.Fatalf("no cambió nada y se avisó a %v", av.quien)
	}
}
