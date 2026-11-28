package negocio

import (
	"context"
	"eventosports/internal/modelo"
	"testing"
	"time"
)

// Reto del taller 13: al crear una inscripción, el servicio avisa «al
// negocio» sin nombrar ningún rol; quién lo recibe lo decide el Avisador.

func (o *operacionFalsa) Activas(uint, uint) (int64, error) { return 0, nil }
func (o *operacionFalsa) CrearInscripcion(i *modelo.Inscripcion) error {
	i.ID = 9
	o.ins = *i
	return nil
}

func TestRetoCrearInscripcionAvisaAlNegocio(t *testing.T) {
	s, _, av := servicioConAvisos("reservada")
	ana := Actor{ID: 1, Rol: "jugador"}
	nueva, err := s.CrearInscripcion(context.Background(), 25, ana)
	if err != nil {
		t.Fatal(err)
	}
	if len(av.negocio) != 1 || av.negocio[0] != (Aviso{Tipo: "inscripcion_nueva", ID: nueva.ID, Estado: "reservada", EventoID: 25}) {
		t.Fatalf("se esperaba un aviso inscripcion_nueva para el negocio y llegó %+v", av.negocio)
	}
	if len(av.quien) != 0 {
		t.Fatalf("crear no le avisa a una persona (tampoco a quien la creó), y se avisó a %v", av.quien)
	}
}

func TestRetoInscripcionRechazadaNoAvisa(t *testing.T) {
	s, op, av := servicioConAvisos("reservada")
	op.ev.Inicio = time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC) // el evento ya empezó
	if _, err := s.CrearInscripcion(context.Background(), 25, Actor{ID: 1, Rol: "jugador"}); err == nil {
		t.Fatal("inscribirse a un evento que ya empezó tenía que fallar")
	}
	if len(av.negocio) != 0 {
		t.Fatalf("no se creó nada y se avisó al negocio: %+v", av.negocio)
	}
}
