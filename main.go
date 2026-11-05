package main

import (
	"log"
	"os"

	"github.com/uleam-web-2026-2/Web2-MENDOZA-QUIJIJE/db"
	"github.com/uleam-web-2026-2/Web2-MENDOZA-QUIJIJE/repository"
)

func main() {
	// La conexión se configura con la misma variable que usa el servidor.
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("falta DATABASE_URL")
	}

	// 1. Inicializar la conexión a la base de datos
	dbConn := db.ConnectDB(dsn)
	defer dbConn.Close()

	// 2. Inicializar el repositorio basado en sqlc
	repo := repository.NewRepository(dbConn)
	_ = repo // Evita error de variable no utilizada temporalmente

	log.Println("¡Servidor y repositorio inicializados correctamente para EventoSports!")
}
