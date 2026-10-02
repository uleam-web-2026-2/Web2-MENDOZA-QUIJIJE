// Package middleware contiene las piezas que se ejecutan en TODAS las peticiones,
// antes y después del handler que realmente atiende la ruta.
//
// ¿Qué es un middleware? Una función que recibe un handler (next) y devuelve
// otro handler que lo envuelve. Esquema:
//
//	func MiMiddleware(next http.Handler) http.Handler {
//	    return http.HandlerFunc(func(w, r) {
//	        // ... código ANTES del handler ...
//	        next.ServeHTTP(w, r)   // ← llama al siguiente eslabón
//	        // ... código DESPUÉS del handler ...
//	    })
//	}
//
// Aquí hay dos:
//   - Registro:     escribe una línea de log por petición (método, ruta, estado, duración).
//   - Recuperacion: si el handler entra en pánico, responde 500 en vez de cortar la conexión.
package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/uleam-web-2026-2/web2A-mendoza-quijije/internal/respuesta"
)

// ---------- Registro ----------

// grabador envuelve al ResponseWriter para saber qué código escribió el handler.
//
// Problema: http.ResponseWriter permite ESCRIBIR el status (WriteHeader) pero no
// LEERLO después. Y Registro necesita saberlo para loguearlo.
// Solución: se le pasa al handler este grabador en lugar del ResponseWriter real.
// El grabador intercepta WriteHeader, anota el código y lo reenvía al original.
type grabador struct {
	http.ResponseWriter     // embebido: Header(), Write(), etc. se heredan tal cual
	estado              int // aquí queda guardado el código que envió el handler
}

// WriteHeader "sobrescribe" el método del ResponseWriter embebido.
func (g *grabador) WriteHeader(codigo int) {
	g.estado = codigo                    // 1. anotar para el log
	g.ResponseWriter.WriteHeader(codigo) // 2. pasar al original para que realmente se envíe
}

// Registro escribe una línea por petición: método, ruta, estado y duración.
// Ejemplo de salida:  GET /tickets/1 → 200 (312µs)
func Registro(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inicio := time.Now() // se toma la hora ANTES de correr el handler

		// 200 por defecto: un handler que escribe el cuerpo sin llamar a WriteHeader
		// envía 200 implícitamente, y en ese caso nuestro WriteHeader nunca se ejecuta.
		g := &grabador{ResponseWriter: w, estado: http.StatusOK}

		next.ServeHTTP(g, r) // ← se pasa g, NO w, para poder espiar el status

		// Va DESPUÉS de next porque el código de estado y la duración
		// solo existen cuando el handler ya terminó de correr.
		log.Printf("%s %s → %d (%v)", r.Method, r.URL.Path, g.estado, time.Since(inicio))
	})
}

// ---------- Recuperacion ----------

// Recuperacion atrapa un pánico y lo convierte en 500 con envoltura,
// en vez de dejar la conexión cortada («Empty reply from server»).
//
// Sin esto, un puntero nulo en cualquier handler mataría la goroutine de esa
// petición y el cliente no recibiría ninguna respuesta HTTP válida.
func Recuperacion(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// defer: esta función se ejecuta SIEMPRE al salir, incluso si hubo pánico.
		// recover() solo funciona dentro de una función diferida; devuelve el valor
		// del pánico si lo hubo, o nil si todo salió bien.
		defer func() {
			if p := recover(); p != nil {
				// El detalle técnico va SOLO al log del servidor. Nunca al cliente:
				// exponer mensajes internos de Go es fuga de información.
				log.Printf("PÁNICO en %s %s: %v", r.Method, r.URL.Path, p)
				// El cliente recibe un 500 genérico con la envoltura de siempre.
				respuesta.Error(w, http.StatusInternalServerError,
					"error_interno", "ocurrió un error inesperado")
			}
		}()
		next.ServeHTTP(w, r) // si esto entra en pánico, el defer de arriba lo atrapa
	})
}
