package main

import (
	"log"
	"os"

	"github.com/uleam-web-2026-2/Web2-MENDOZA-QUIJIJE/db"
	"github.com/uleam-web-2026-2/Web2-MENDOZA-QUIJIJE/repository"
)

func main() {
	// Cadena de conexión a PostgreSQL (puedes ajustarla según tus credenciales locales)
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:1234567@postgres:5432/eventosports?sslmode=disable"
	}

	// 1. Inicializar la conexión a la base de datos
	dbConn := db.ConnectDB(dsn)
	defer dbConn.Close()

	// 2. Inicializar el repositorio basado en sqlc
	repo := repository.NewRepository(dbConn)
	_ = repo // Evita error de variable no utilizada temporalmente

	log.Println("¡Servidor y repositorio inicializados correctamente para EventoSports!")
}
