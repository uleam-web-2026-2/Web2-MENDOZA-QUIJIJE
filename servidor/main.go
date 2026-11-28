package main

import (
	"eventosports/internal/configuracion"
	"eventosports/internal/datos"
	"eventosports/internal/eventosports"
	"eventosports/internal/negocio"
	"eventosports/internal/sesion"
	"flag"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"net/http"
	"time"
)

// origenFront es el único front que puede leer las respuestas (CORS) y
// abrir el WebSocket. Se escribe una vez, igual que en la barra del navegador.
const origenFront = "http://localhost:5173"

func main() {
	migrar := flag.Bool("migrar", false, "con el dueño de las tablas: migra, siembra y da permisos al usuario de la API; después termina")
	flag.Parse()
	c, err := configuracion.Leer()
	if err != nil {
		log.Fatal(err)
	}
	cfg := &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)}
	if *migrar {
		if err := migrarConDueno(c, cfg); err != nil {
			log.Fatal("migración: ", err)
		}
		log.Println("migración lista: tablas, ejemplos y permisos del usuario de la API")
		return
	}
	// La API se conecta con su propio usuario (DATABASE_URL), que no es
	// superusuario ni dueño de las tablas. Por eso ya no corre AutoMigrate:
	// las tablas las crea y las cambia el dueño con go run . -migrar.
	db, err := gorm.Open(postgres.Open(c.Base), cfg)
	if err != nil {
		log.Fatal("no se pudo conectar a PostgreSQL; revise DATABASE_URL y el servicio")
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	defer sqlDB.Close()
	r := chi.NewRouter()
	// CORS arriba de todo: responde las peticiones previas (OPTIONS), que
	// no tienen manejador ni token. Mínimo privilegio: se nombra el front,
	// no "*", y solo los métodos que usa la API (no hay PUT ni DELETE).
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{origenFront},
		AllowedMethods: []string{"GET", "POST", "PATCH"},
		AllowedHeaders: []string{"Authorization", "Content-Type"},
		MaxAge:         300, // segundos que el navegador recuerda la respuesta
	}))
	r.Use(middleware.ClientIPFromRemoteAddr) // anota la IP del cliente (la de la conexión)
	// Registro de cada petición en la terminal (debajo de CORS: las previas no se anotan).
	r.Use(middleware.Logger)
	tokens := sesion.NuevosTokens(c.Secreto)
	servicio := &negocio.Servicio{Almacen: &datos.AlmacenGORM{DB: db}}
	// Los avisos: el servicio avisa al jugador cuando cambia su inscripción.
	// GET /ws queda en Rutas, fuera del guardia (el token va en la dirección).
	avisos := eventosports.NuevosAvisos(tokens, origenFront)
	servicio.AvisarCon(avisos)
	// Login: 5 intentos por minuto por IP; el sexto recibe 429 con Retry-After.
	limiteLogin := httprate.LimitBy(5, time.Minute, porIP)
	(&eventosports.Manejador{Servicio: servicio, Sesion: &sesion.Manejador{DB: db, Tokens: tokens}, LimiteLogin: limiteLogin, Avisos: avisos}).Rutas(r, tokens)
	servidor := &http.Server{Addr: ":" + c.Puerto, Handler: r, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second}
	log.Printf("EventoSports listo en http://localhost:%s", c.Puerto)
	log.Fatal(servidor.ListenAndServe())
}

// porIP dice a quién se le cuenta cada petición: a la IP que anotó
// ClientIPFromRemoteAddr, la de la conexión. No se lee ninguna cabecera
// que el cliente pueda escribir (como X-Forwarded-For).
func porIP(r *http.Request) (string, error) {
	ip := middleware.GetClientIP(r.Context())
	return httprate.CanonicalizeIP(ip), nil
}
