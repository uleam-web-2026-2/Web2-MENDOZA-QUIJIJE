// Package respuesta define el formato ÚNICO en que el servidor responde.
//
// Toda respuesta, sea éxito o error, tiene la misma forma (la "envoltura"):
//
//	{ "ok": true,  "datos": { ... } }                                   ← éxito
//	{ "ok": false, "error": { "codigo": "...", "mensaje": "..." } }    ← error
//
// Ventaja: el cliente siempre sabe dónde buscar. Mira "ok"; si es true lee
// "datos", si es false lee "error.codigo". Nunca tiene que adivinar.
//
// Solo se exportan dos funciones: Exito y Error. Todo lo demás es privado
// (minúscula) para que ningún handler pueda inventar otro formato.
package respuesta

import (
	"encoding/json"
	"net/http"
)

// detalleError es lo que va dentro de "error" cuando algo falla.
//   - Codigo: palabra fija y estable (no_encontrado, datos_invalidos...) para que
//     el cliente pueda programar con ella: if codigo == "no_encontrado" { ... }.
//   - Mensaje: texto para humanos. Puede cambiar sin romper a los clientes.
type detalleError struct {
	Codigo  string `json:"codigo"`
	Mensaje string `json:"mensaje"`
}

// envoltura es la estructura completa que se serializa a JSON.
// Las etiquetas `json:"..."` definen el nombre de cada campo en el JSON.
type envoltura struct {
	OK    bool          `json:"ok"`
	Datos any           `json:"datos,omitempty"` // omitempty: si está vacío, el campo no aparece en el JSON
	Error *detalleError `json:"error,omitempty"` // puntero: nil = "no hay error" → se omite
}

// escribir es el único lugar donde realmente se escribe en la respuesta HTTP.
// Los tres pasos van en este orden obligatorio:
func escribir(w http.ResponseWriter, estado int, cuerpo envoltura) {
	// 1. Cabeceras: hay que fijarlas ANTES de WriteHeader, después ya no se pueden tocar.
	w.Header().Set("Content-Type", "application/json; charset=utf8")
	// 2. Código de estado (200, 404, 500...). Punto de no retorno: una vez enviado
	//    no se puede cambiar; por eso los handlers validan todo antes de llegar aquí.
	w.WriteHeader(estado)
	// 3. Cuerpo: el struct se convierte a JSON y se escribe directo en la conexión.
	//    Si falla aquí ya es tarde para avisar al cliente (la cabecera ya salió),
	//    por eso se ignora el error con "_ =".
	_ = json.NewEncoder(w).Encode(cuerpo)
}

// Exito responde con el estado indicado (200, 201...) y los datos dentro de la envoltura.
// Ejemplo: Exito(w, http.StatusCreated, ticket) → 201 {"ok":true,"datos":{...}}
func Exito(w http.ResponseWriter, estado int, datos any) {
	escribir(w, estado, envoltura{OK: true, Datos: datos})
}

// Error responde con la envoltura de error. Nunca expone detalles internos
// (trazas, mensajes de Go, rutas de archivos): solo el código y un mensaje pensado
// para el cliente. El detalle técnico, si hace falta, va al log del servidor.
// Ejemplo: Error(w, 404, "no_encontrado", "no existe un ticket con ese id")
func Error(w http.ResponseWriter, estado int, codigo, mensaje string) {
	escribir(w, estado, envoltura{OK: false,
		Error: &detalleError{Codigo: codigo, Mensaje: mensaje}})
}
