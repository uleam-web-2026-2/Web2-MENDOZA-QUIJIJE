// EventoSports · Hito 1: servidor mínimo con los seis endpoints de la sección 7 de la ficha.
// Los datos viven en memoria (se pierden al apagar el servidor); la conexión a PostgreSQL con GORM es un paso posterior.
package main

import (
	"log"
	"net/http"
	"os"

	"eventosports/internal/eventosports"
)

func main() {
	puerto := os.Getenv("PORT")
	if puerto == "" {
		puerto = "8080"
	}
	log.Printf("EventoSports escuchando en http://localhost:%s", puerto)
	log.Fatal(http.ListenAndServe(":"+puerto, eventosports.NuevoServidor().Rutas()))
}
