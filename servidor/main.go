package main

import (
	"eventosports/internal/configuracion"
	"eventosports/internal/datos"
	"eventosports/internal/eventosports"
	"eventosports/internal/negocio"
	"eventosports/internal/sesion"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
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
	if err := datos.Migrar(db); err != nil {
		log.Fatal("no se pudieron crear las tablas: ", err)
	}
	if c.Demo {
		if err := datos.SembrarDemo(db); err != nil {
			log.Fatal("no se pudieron preparar los ejemplos: ", err)
		}
	}
	r := chi.NewRouter()
	// CORS arriba de todo: responde las peticiones previas (OPTIONS), que
	// no tienen manejador ni token. Mínimo privilegio: se nombra el front,
	// no "*", y solo los métodos que usa la API (no hay PUT ni DELETE).
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"http://localhost:5173"},
		AllowedMethods: []string{"GET", "POST", "PATCH"},
		AllowedHeaders: []string{"Authorization", "Content-Type"},
		MaxAge:         300, // segundos que el navegador recuerda la respuesta
	}))
	// Registro de cada petición en la terminal (debajo de CORS: las previas no se anotan).
	r.Use(middleware.Logger)
	tokens := sesion.NuevosTokens(c.Secreto)
	servicio := &negocio.Servicio{Almacen: &datos.AlmacenGORM{DB: db}}
	(&eventosports.Manejador{Servicio: servicio, Sesion: &sesion.Manejador{DB: db, Tokens: tokens}}).Rutas(r, tokens)
	servidor := &http.Server{Addr: ":" + c.Puerto, Handler: r, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second}
	log.Printf("EventoSports listo en http://localhost:%s", c.Puerto)
	log.Fatal(servidor.ListenAndServe())
}
