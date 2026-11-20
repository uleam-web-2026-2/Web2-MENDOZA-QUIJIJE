package eventosports

import (
	"eventosports/internal/datos"
	"eventosports/internal/negocio"
	"eventosports/internal/sesion"
	"net/http"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
)

// Semana 12 · punto 3. El límite va solo delante de POST /login: cinco
// intentos fallidos dan 401 y el sexto 429 con Retry-After, sin llegar a
// bcrypt. Las demás rutas siguen respondiendo aunque el login esté frenado.
func TestLimiteSoloEnElLogin(t *testing.T) {
	db := baseDePrueba(t)
	r := chi.NewRouter()
	m := &Manejador{
		Servicio:    &negocio.Servicio{Almacen: &datos.AlmacenGORM{DB: db}},
		Sesion:      &sesion.Manejador{DB: db, Tokens: tokensDePrueba},
		LimiteLogin: httprate.LimitBy(5, time.Minute, func(*http.Request) (string, error) { return "una-ip", nil }),
	}
	m.Rutas(r, tokensDePrueba)

	malo := `{"usuario":"ana","clave":"no-es-esta"}`
	for i := 1; i <= 5; i++ {
		esperar(t, llamar(r, "POST", "/login", malo, ""), 401)
	}
	sexto := llamar(r, "POST", "/login", malo, "")
	esperar(t, sexto, 429)
	if sexto.Header().Get("Retry-After") == "" {
		t.Fatal("el 429 debía traer Retry-After")
	}
	esperar(t, llamar(r, "GET", "/eventos", "", ""), 200)
}
