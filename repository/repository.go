package repository

import (
	"context"
	"database/sql"

	"github.com/uleam-web-2026-2/Web2-MENDOZA-QUIJIJE/db"
)

// Repository define la interfaz abstracta para interactuar con la base de datos
type Repository interface {
	ListDeportes(ctx context.Context) ([]db.Deporte, error)
	GetEvento(ctx context.Context, id int32) (db.Evento, error)
	ListEventos(ctx context.Context) ([]db.Evento, error)
	CreateEvento(ctx context.Context, arg db.CreateEventoParams) (db.Evento, error)
	UpdateEvento(ctx context.Context, arg db.UpdateEventoParams) (db.Evento, error)
	DeleteEvento(ctx context.Context, id int32) error
}

// SQLRepository implementa la interfaz Repository usando sqlc
type SQLRepository struct {
	*db.Queries
	dbConn *sql.DB
}

// NewRepository inicializa un nuevo repositorio con la conexión a la BD
func NewRepository(dbConn *sql.DB) Repository {
	return &SQLRepository{
		Queries: db.New(dbConn),
		dbConn:  dbConn,
	}
}
