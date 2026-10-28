// Package negocio contiene casos de uso y reglas sin HTTP ni GORM.
package negocio

import (
	"context"
	"errors"
	"eventosports/internal/modelo"
	"time"
)

type TipoError string

const (
	Invalido  TipoError = "invalido"
	Prohibido TipoError = "prohibido"
	Conflicto TipoError = "conflicto"
	Regla     TipoError = "regla"
)

type Error struct {
	Tipo    TipoError
	Mensaje string
}

func (e *Error) Error() string                      { return e.Mensaje }
func rechazar(tipo TipoError, mensaje string) error { return &Error{tipo, mensaje} }

var ErrNoExiste = errors.New("el recurso no existe")

type Actor struct {
	ID  uint
	Rol string
}
type Filtro struct{ Deporte, Nivel, Ciudad string }
type CrearEvento struct {
	Titulo  string    `json:"titulo"`
	Deporte string    `json:"deporte"`
	Nivel   string    `json:"nivel"`
	Inicio  time.Time `json:"inicio"`
	Cupos   int       `json:"cupos"`
	Precio  float64   `json:"precio"`
	SedeID  uint      `json:"sede_id"`
}
type CambioEstado struct {
	Estado     string `json:"estado"`
	Banco      string `json:"banco"`
	Referencia string `json:"referencia"`
}
type EventoVista struct {
	modelo.Evento
	Sede          modelo.Sede         `json:"sede"`
	CuposLibres   int                 `json:"cupos_libres"`
	MiInscripcion *modelo.Inscripcion `json:"mi_inscripcion,omitempty"`
}
type JugadorVista struct {
	ID     uint   `json:"id"`
	Nombre string `json:"nombre"`
}
type InscripcionVista struct {
	modelo.Inscripcion
	Jugador    JugadorVista `json:"jugador"`
	UltimoPago *modelo.Pago `json:"ultimo_pago,omitempty"`
}

// Las interfaces pertenecen a quien las consume: el negocio decide que necesita.
type Consulta interface {
	Eventos(Filtro) ([]modelo.Evento, error)
	Evento(uint) (modelo.Evento, error)
	Sede(uint) (modelo.Sede, error)
	Usuario(uint) (modelo.Usuario, error)
	Inscripcion(uint) (modelo.Inscripcion, error)
	Confirmadas(uint) (int64, error)
	Inscripciones(uint) ([]modelo.Inscripcion, error)
	UltimaInscripcion(uint, uint) (modelo.Inscripcion, error)
	UltimoPago(uint) (modelo.Pago, error)
	CrearEvento(*modelo.Evento) error
}
type Operacion interface {
	Inscripcion(uint) (modelo.Inscripcion, error)
	EventoBloqueado(uint) (modelo.Evento, error)
	InscripcionBloqueada(uint) (modelo.Inscripcion, error)
	Activas(uint, uint) (int64, error)
	Confirmadas(uint) (int64, error)
	CrearInscripcion(*modelo.Inscripcion) error
	CrearPago(*modelo.Pago) error
	UltimoPago(uint) (modelo.Pago, error)
	VerificarPago(uint) error
	GuardarEstado(modelo.Inscripcion) error
}
type Almacen interface {
	ConContexto(context.Context) Consulta
	Transaccion(context.Context, func(Operacion) error) error
}
type Servicio struct {
	Almacen Almacen
	Ahora   func() time.Time
}

func (s *Servicio) reloj() time.Time {
	if s.Ahora != nil {
		return s.Ahora()
	}
	return time.Now()
}
