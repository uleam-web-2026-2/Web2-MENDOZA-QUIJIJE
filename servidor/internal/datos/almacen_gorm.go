// Package datos adapta las interfaces del negocio a GORM y PostgreSQL.
package datos

import (
	"context"
	"errors"
	"eventosports/internal/modelo"
	"eventosports/internal/negocio"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AlmacenGORM struct{ DB *gorm.DB }
type consulta struct{ db *gorm.DB }

var _ negocio.Almacen = (*AlmacenGORM)(nil)
var _ negocio.Consulta = (*consulta)(nil)
var _ negocio.Operacion = (*consulta)(nil)

func normalizar(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return negocio.ErrNoExiste
	}
	return err
}
func (a *AlmacenGORM) ConContexto(ctx context.Context) negocio.Consulta {
	return &consulta{a.DB.WithContext(ctx)}
}
func (a *AlmacenGORM) Transaccion(ctx context.Context, f func(negocio.Operacion) error) error {
	return normalizar(a.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error { return f(&consulta{tx}) }))
}
func (q *consulta) Eventos(f negocio.Filtro) ([]modelo.Evento, error) {
	db := q.db.Model(&modelo.Evento{})
	for _, campo := range []struct{ nombre, valor string }{{"deporte", f.Deporte}, {"nivel", f.Nivel}} {
		if campo.valor != "" {
			db = db.Where(campo.nombre+" = ?", campo.valor)
		}
	}
	if f.Ciudad != "" {
		db = db.Where("sede_id IN (?)", q.db.Model(&modelo.Sede{}).Select("id").Where("LOWER(ciudad) = LOWER(?)", f.Ciudad))
	}
	var v []modelo.Evento
	err := db.Order("id").Find(&v).Error
	return v, normalizar(err)
}
func (q *consulta) Evento(id uint) (modelo.Evento, error) {
	var v modelo.Evento
	err := q.db.First(&v, id).Error
	return v, normalizar(err)
}
func (q *consulta) Sede(id uint) (modelo.Sede, error) {
	var v modelo.Sede
	err := q.db.First(&v, id).Error
	return v, normalizar(err)
}
func (q *consulta) Usuario(id uint) (modelo.Usuario, error) {
	var v modelo.Usuario
	err := q.db.First(&v, id).Error
	return v, normalizar(err)
}
func (q *consulta) Inscripcion(id uint) (modelo.Inscripcion, error) {
	var v modelo.Inscripcion
	err := q.db.First(&v, id).Error
	return v, normalizar(err)
}
func (q *consulta) Confirmadas(id uint) (int64, error) {
	var n int64
	err := q.db.Model(&modelo.Inscripcion{}).Where("evento_id = ? AND estado = ?", id, "confirmada").Count(&n).Error
	return n, normalizar(err)
}
func (q *consulta) Inscripciones(id uint) ([]modelo.Inscripcion, error) {
	var v []modelo.Inscripcion
	err := q.db.Where("evento_id = ?", id).Order("id").Find(&v).Error
	return v, normalizar(err)
}
func (q *consulta) UltimaInscripcion(evento, jugador uint) (modelo.Inscripcion, error) {
	var v modelo.Inscripcion
	err := q.db.Where("evento_id = ? AND jugador_id = ?", evento, jugador).Order("id DESC").First(&v).Error
	return v, normalizar(err)
}
func (q *consulta) UltimoPago(id uint) (modelo.Pago, error) {
	var v modelo.Pago
	err := q.db.Where("inscripcion_id = ?", id).Order("id DESC").First(&v).Error
	return v, normalizar(err)
}
func (q *consulta) CrearEvento(v *modelo.Evento) error { return normalizar(q.db.Create(v).Error) }
func (q *consulta) EventoBloqueado(id uint) (modelo.Evento, error) {
	var v modelo.Evento
	err := q.db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&v, id).Error
	return v, normalizar(err)
}
func (q *consulta) InscripcionBloqueada(id uint) (modelo.Inscripcion, error) {
	var v modelo.Inscripcion
	err := q.db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&v, id).Error
	return v, normalizar(err)
}
func (q *consulta) Activas(evento, jugador uint) (int64, error) {
	var n int64
	err := q.db.Model(&modelo.Inscripcion{}).Where("evento_id = ? AND jugador_id = ? AND estado <> ?", evento, jugador, "cancelada").Count(&n).Error
	return n, normalizar(err)
}
func (q *consulta) CrearInscripcion(v *modelo.Inscripcion) error {
	return normalizar(q.db.Create(v).Error)
}
func (q *consulta) CrearPago(v *modelo.Pago) error { return normalizar(q.db.Create(v).Error) }
func (q *consulta) VerificarPago(id uint) error {
	return normalizar(q.db.Model(&modelo.Pago{}).Where("id = ?", id).Update("verificado", true).Error)
}
func (q *consulta) GuardarEstado(v modelo.Inscripcion) error {
	return normalizar(q.db.Model(&v).Update("estado", v.Estado).Error)
}
