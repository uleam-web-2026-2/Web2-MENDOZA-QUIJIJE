package db

import (
	"database/sql"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// ConnectDB abre y verifica la conexión a PostgreSQL
func ConnectDB(dataSourceName string) *sql.DB {
	conn, err := sql.Open("pgx", dataSourceName)
	if err != nil {
		log.Fatalf("Error al abrir la base de datos: %v", err)
	}

	if err = conn.Ping(); err != nil {
		log.Fatalf("Error al conectar (ping) a PostgreSQL: %v", err)
	}

	log.Println("¡Conexión exitosa a PostgreSQL con pgx!")
	return conn
}
