package eventosports

import (
	"context"
	"eventosports/internal/negocio"
	"eventosports/internal/sesion"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/go-chi/chi/v5"
)

const origenDePrueba = "http://localhost:5173"

// Los usuarios de la semilla: ana (1) y eva (3) son jugadoras, luis (2) organizador.
var (
	ana  = sesion.Usuario{ID: 1, Nombre: "ana", Rol: sesion.RolJugador}
	eva  = sesion.Usuario{ID: 3, Nombre: "eva", Rol: sesion.RolJugador}
	luis = sesion.Usuario{ID: 2, Nombre: "luis", Rol: sesion.RolOrganizador}
)

// servidorDeAvisos levanta un servidor de verdad, en un puerto libre, con
// las mismas Rutas de main.go (y así /ws queda donde queda en main.go).
// Devuelve su dirección ws:// y los avisos, para enviar.
func servidorDeAvisos(t *testing.T) (string, *Avisos) {
	t.Helper()
	avisos := NuevosAvisos(tokensDePrueba, origenDePrueba)
	r := chi.NewRouter()
	(&Manejador{Servicio: &negocio.Servicio{}, Sesion: &sesion.Manejador{Tokens: tokensDePrueba}, Avisos: avisos}).Rutas(r, tokensDePrueba)
	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)
	return "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws", avisos
}

// conectar abre el WebSocket como lo abriría el front: con el token en la
// dirección y el Origin del front. Con un usuario vacío, va sin token.
func conectar(t *testing.T, url string, u sesion.Usuario, origen string) (*websocket.Conn, *http.Response, error) {
	t.Helper()
	if u.ID != 0 {
		token, err := tokensDePrueba.Emitir(u)
		if err != nil {
			t.Fatal(err)
		}
		url += "?token=" + token
	}
	ctx, cancelar := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelar()
	cabeceras := http.Header{}
	if origen != "" {
		cabeceras.Set("Origin", origen)
	}
	return websocket.Dial(ctx, url, &websocket.DialOptions{HTTPHeader: cabeceras})
}

// noLlega comprueba que en medio segundo no le llegue nada a esa conexión.
func noLlega(t *testing.T, c *websocket.Conn, quien string) {
	t.Helper()
	ctx, cancelar := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancelar()
	var ajeno negocio.Aviso
	if err := wsjson.Read(ctx, c, &ajeno); err == nil {
		t.Fatalf("a %s le llegó un aviso que no le toca: %+v", quien, ajeno)
	}
}

func TestWebSocketSinTokenResponde401(t *testing.T) {
	url, _ := servidorDeAvisos(t)
	_, resp, err := conectar(t, url, sesion.Usuario{}, origenDePrueba)
	if err == nil || resp == nil || resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("se esperaba 401 y llegó %v (err %v)", resp, err)
	}
}

func TestWebSocketDesdeOtroOrigenResponde403(t *testing.T) {
	url, _ := servidorDeAvisos(t)
	_, resp, err := conectar(t, url, ana, "http://malo.example")
	if err == nil || resp == nil || resp.StatusCode != http.StatusForbidden {
		t.Fatalf("se esperaba 403 y llegó %v (err %v)", resp, err)
	}
}

// Con un token vencido o cambiado también es 401, y el mensaje es el de
// Conectar, no el del guardia: /ws está fuera del grupo de Autenticar.
func TestWebSocketFueraDelGuardia(t *testing.T) {
	r := chi.NewRouter()
	(&Manejador{Servicio: &negocio.Servicio{}, Sesion: &sesion.Manejador{Tokens: tokensDePrueba}, Avisos: NuevosAvisos(tokensDePrueba, origenDePrueba)}).Rutas(r, tokensDePrueba)
	rec := llamar(r, "GET", "/ws?token=basura", "", "")
	esperar(t, rec, http.StatusUnauthorized)
	if !strings.Contains(rec.Body.String(), "sin sesión") {
		t.Fatalf("respondió el guardia, no Conectar: %s", rec.Body.String())
	}
}

func TestElAvisoLlegaSoloAlJugadorDueno(t *testing.T) {
	url, avisos := servidorDeAvisos(t)
	connAna, _, err := conectar(t, url, ana, origenDePrueba)
	if err != nil {
		t.Fatal(err)
	}
	defer connAna.CloseNow()
	connEva, _, err := conectar(t, url, eva, origenDePrueba)
	if err != nil {
		t.Fatal(err)
	}
	defer connEva.CloseNow()
	time.Sleep(100 * time.Millisecond) // que las dos queden suscritas

	avisos.Enviar(ana.ID, negocio.Aviso{Tipo: "estado_cambiado", ID: 7, Estado: "confirmada"})

	// A ana le llega, con tipo e id.
	ctx, cancelar := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelar()
	var recibido negocio.Aviso
	if err := wsjson.Read(ctx, connAna, &recibido); err != nil {
		t.Fatalf("a ana no le llegó el aviso: %v", err)
	}
	if recibido.ID != 7 || recibido.Tipo != "estado_cambiado" {
		t.Fatalf("a ana le llegó %+v", recibido)
	}
	// A eva, del mismo rol, en medio segundo no le llega nada.
	noLlega(t, connEva, "eva")
}
