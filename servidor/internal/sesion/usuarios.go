package sesion

// En este archivo cambian SOLO las líneas marcadas con «CAMBIEN»:
// los dos roles de su negocio y, si quieren, los dos usuarios de ejemplo.

import (
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Los dos roles de su negocio, escritos como en la sección 6 de su ficha.
// El nombre de la constante empieza con Rol; el texto va en minúsculas,
// sin espacios y sin tildes.
const (
	RolJugador     = "jugador"     // CAMBIEN: el rol que pide o reporta
	RolOrganizador = "organizador" // CAMBIEN: el rol que atiende o decide
)

// Usuario es una persona que puede iniciar sesión.
// De la contraseña se guarda un resumen (bcrypt), nunca la contraseña.
type Usuario struct {
	ID        uint   `json:"id"`
	Nombre    string `json:"nombre" gorm:"uniqueIndex"`
	Rol       string `json:"rol"`
	ClaveHash string `json:"-"`
}

// Solo si su modelo YA tiene otra tabla llamada «usuarios»: quiten las
// dos barras de las tres líneas siguientes, y la tabla de este archivo
// pasa a llamarse «cuentas».
func (Usuario) TableName() string {
	return "cuentas"
}

// Sembrar deja dos usuarios de ejemplo, uno por rol. Corre en cada
// arranque: si cambian un rol o una contraseña aquí, al reiniciar el
// servidor el usuario queda actualizado. Son usuarios y contraseñas
// INVENTADOS, solo para probar: nunca una contraseña real.
func Sembrar(db *gorm.DB) error {
	ejemplos := []struct{ nombre, rol, clave string }{
		{"ana", RolJugador, "ana-2026"},       // CAMBIEN el rol
		{"luis", RolOrganizador, "luis-2026"}, // CAMBIEN el rol
	}
	for _, e := range ejemplos {
		// bcrypt calcula el resumen: eso es lo que se guarda.
		resumen, err := bcrypt.GenerateFromPassword(
			[]byte(e.clave), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		var u Usuario
		db.Where("nombre = ?", e.nombre).Limit(1).Find(&u)
		u.Nombre, u.Rol, u.ClaveHash = e.nombre, e.rol, string(resumen)
		if err := db.Save(&u).Error; err != nil {
			return err
		}
	}
	return nil
}

// Manejador tiene la única ruta de este paquete: iniciar sesión.
type Manejador struct {
	DB     *gorm.DB
	Tokens *Tokens
}

// Login: POST /login con {"usuario": "...", "clave": "..."}.
// 400 si el JSON está roto · 401 si no coincide · 200 con el token.
func (m *Manejador) Login(w http.ResponseWriter, r *http.Request) {
	var pedido struct {
		Usuario string `json:"usuario"`
		Clave   string `json:"clave"`
	}
	if err := json.NewDecoder(r.Body).Decode(&pedido); err != nil {
		responder(w, http.StatusBadRequest, "el cuerpo no es JSON válido")
		return
	}
	// El mismo mensaje para «no existe» y para «contraseña incorrecta»:
	// quien prueba nombres al azar no averigua cuáles existen.
	const noCoincide = "usuario o contraseña incorrectos"
	// ---- desde aquí hace falta la base de datos ----
	var u Usuario
	err := m.DB.Where("nombre = ?", pedido.Usuario).Limit(1).Find(&u).Error
	if err != nil {
		responder(w, http.StatusInternalServerError, "no se pudo leer")
		return
	}
	// bcrypt compara la contraseña que llegó con el resumen guardado.
	// Si el usuario no existe, u.ID es 0 y no hay nada que comparar.
	if u.ID == 0 || bcrypt.CompareHashAndPassword(
		[]byte(u.ClaveHash), []byte(pedido.Clave)) != nil {
		responder(w, http.StatusUnauthorized, noCoincide)
		return
	}
	token, err := m.Tokens.Emitir(u)
	if err != nil {
		responder(w, http.StatusInternalServerError,
			"no se pudo emitir el token")
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
