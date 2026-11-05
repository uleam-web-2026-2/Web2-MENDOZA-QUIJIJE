package eventosports

import (
	"time"

	"eventosports/internal/modelo"
	"eventosports/internal/negocio"
)

type cambioEstadoPeticion struct {
	Estado     string `json:"estado"`
	Banco      string `json:"banco"`
	Referencia string `json:"referencia"`
}

type inscripcionRespuesta struct {
	ID        uint      `json:"id"`
	EventoID  uint      `json:"evento_id"`
	JugadorID uint      `json:"jugador_id"`
	Estado    string    `json:"estado"`
	Creado    time.Time `json:"creado"`
}

func nuevaInscripcionRespuesta(i modelo.Inscripcion) inscripcionRespuesta {
	return inscripcionRespuesta{
		ID: i.ID, EventoID: i.EventoID, JugadorID: i.JugadorID,
		Estado: i.Estado, Creado: i.Creado,
	}
}

type eventoRespuesta struct {
	modelo.Evento
	Sede          modelo.Sede           `json:"sede"`
	CuposLibres   int                   `json:"cupos_libres"`
	MiInscripcion *inscripcionRespuesta `json:"mi_inscripcion,omitempty"`
}

func nuevoEventoRespuesta(v negocio.EventoVista) eventoRespuesta {
	respuesta := eventoRespuesta{Evento: v.Evento, Sede: v.Sede, CuposLibres: v.CuposLibres}
	if v.MiInscripcion != nil {
		i := nuevaInscripcionRespuesta(*v.MiInscripcion)
		respuesta.MiInscripcion = &i
	}
	return respuesta
}

type inscripcionVistaRespuesta struct {
	inscripcionRespuesta
	Jugador    negocio.JugadorVista `json:"jugador"`
	UltimoPago *modelo.Pago         `json:"ultimo_pago,omitempty"`
}
