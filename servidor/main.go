package main

import (
	"eventosports/internal/configuracion"
	"eventosports/internal/eventosports"
	"eventosports/internal/sesion"
	"github.com/go-chi/chi/v5"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"net/http"
	"time"
)

func main() {
	c, err := configuracion.Leer()
	if err != nil {
		log.Fatal(err)
	}
	db, err := gorm.Open(postgres.Open(c.Base), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		log.Fatal("no se pudo conectar a PostgreSQL; revise DATABASE_URL y el servicio")
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	defer sqlDB.Close()
	if err := eventosports.Migrar(db); err != nil {
		log.Fatal("no se pudieron crear las tablas: ", err)
	}
	if c.Demo {
		if err := eventosports.SembrarDemo(db); err != nil {
			log.Fatal("no se pudieron preparar los ejemplos: ", err)
		}
	}
	r := chi.NewRouter()
	(&eventosports.Manejador{DB: db}).Rutas(r, sesion.NuevosTokens(c.Secreto))
	servidor := &http.Server{Addr: ":" + c.Puerto, Handler: r, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second}
	log.Printf("EventoSports listo en http://localhost:%s", c.Puerto)
	log.Fatal(servidor.ListenAndServe())
}
