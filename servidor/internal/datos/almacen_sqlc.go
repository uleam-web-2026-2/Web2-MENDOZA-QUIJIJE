package datos

import (
	"context"
	"database/sql"
	"errors"

	"eventosports/internal/datos/filas"
	"eventosports/internal/modelo"
	"eventosports/internal/negocio"

	"gorm.io/gorm"
)

type AlmacenSQLC struct{ *AlmacenGORM }

type consultaSQLC struct {
	*consulta
	ctx context.Context
	q   *filas.Queries
}

var _ negocio.Almacen = (*AlmacenSQLC)(nil)
var _ negocio.Operacion = (*consultaSQLC)(nil)

func (a *AlmacenSQLC) ConContexto(ctx context.Context) negocio.Consulta {
	return nuevaConsultaSQLC(ctx, a.DB.WithContext(ctx))
}

func (a *AlmacenSQLC) Transaccion(ctx context.Context, f func(negocio.Operacion) error) error {
	return a.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return f(nuevaConsultaSQLC(ctx, tx))
	})
}

func nuevaConsultaSQLC(ctx context.Context, db *gorm.DB) *consultaSQLC {
	// ConnPool es la conexión activa o la transacción abierta por GORM.
	return &consultaSQLC{consulta: &consulta{db}, ctx: ctx, q: filas.New(db.Statement.ConnPool)}
}

func (q *consultaSQLC) Inscripcion(id uint) (modelo.Inscripcion, error) {
	f, err := q.q.BuscarInscripcion(q.ctx, int64(id))
	if errors.Is(err, sql.ErrNoRows) {
		return modelo.Inscripcion{}, negocio.ErrNoExiste
	}
	if err != nil {
		return modelo.Inscripcion{}, err
	}
	return modelo.Inscripcion{
		ID: uint(f.ID), EventoID: uint(f.EventoID), JugadorID: uint(f.JugadorID),
		Estado: f.Estado, Creado: f.Creado.Time,
	}, nil
}

func (q *consultaSQLC) GuardarEstado(i modelo.Inscripcion) error {
	n, err := q.q.GuardarEstadoInscripcion(q.ctx, filas.GuardarEstadoInscripcionParams{
		ID: int64(i.ID), Estado: i.Estado,
	})
	if err == nil && n == 0 {
		return negocio.ErrNoExiste
	}
	return err
}
