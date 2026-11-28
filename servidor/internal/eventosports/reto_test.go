package eventosports

import (
	"context"
	"eventosports/internal/modelo"
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

// Reto del taller 13: al crear una inscripción, los organizadores
// conectados reciben un aviso; los jugadores no, tampoco quien la creó.

// marta es una segunda organizadora, solo de esta prueba.
var marta = sesion.Usuario{ID: 4, Nombre: "marta", Rol: sesion.RolOrganizador}

// almacenReto cumple negocio.Almacen sin base de datos, con lo justo para
// que CrearInscripcion funcione: el evento 31 de luis, con cupos y futuro.
type almacenReto struct {
	negocio.Almacen
	ev modelo.Evento
}

func (a *almacenReto) Transaccion(_ context.Context, f func(negocio.Operacion) error) error {
	return f(&operacionReto{ev: a.ev})
}

type operacionReto struct {
	negocio.Operacion
	ev modelo.Evento
}

func (o *operacionReto) EventoBloqueado(uint) (modelo.Evento, error)  { return o.ev, nil }
func (o *operacionReto) Activas(uint, uint) (int64, error)            { return 0, nil }
func (o *operacionReto) Confirmadas(uint) (int64, error)              { return 0, nil }
func (o *operacionReto) CrearInscripcion(i *modelo.Inscripcion) error { i.ID = 9; return nil }

// servidorDelReto levanta la API de verdad (Rutas de main.go, el servicio
// con AvisarCon y /ws) sobre el almacén falso. Devuelve la dirección ws://,
// la http:// y los avisos.
func servidorDelReto(t *testing.T) (string, string, *Avisos) {
	t.Helper()
	avisos := NuevosAvisos(tokensDePrueba, origenDePrueba, sesion.RolOrganizador) // igual que main.go
	servicio := &negocio.Servicio{Almacen: &almacenReto{ev: modelo.Evento{ID: 31, OrganizadorID: luis.ID, Cupos: 14, Precio: 2.5, Inicio: time.Now().Add(10 * 24 * time.Hour)}}}
	servicio.AvisarCon(avisos)
	r := chi.NewRouter()
	(&Manejador{Servicio: servicio, Sesion: &sesion.Manejador{Tokens: tokensDePrueba}, Avisos: avisos}).Rutas(r, tokensDePrueba)
	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)
	return "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws", srv.URL, avisos
}

// conectarATodos abre una conexión por usuario: ana y eva (jugadoras),
// luis y marta (organizadores).
func conectarATodos(t *testing.T, url string) map[string]*websocket.Conn {
	t.Helper()
	conns := map[string]*websocket.Conn{}
	for _, u := range []sesion.Usuario{ana, eva, luis, marta} {
		c, _, err := conectar(t, url, u, origenDePrueba)
		if err != nil {
			t.Fatalf("%s no pudo conectarse: %v", u.Nombre, err)
		}
		t.Cleanup(func() { c.CloseNow() })
		conns[u.Nombre] = c
	}
	time.Sleep(100 * time.Millisecond) // que todos queden suscritos
	return conns
}

func leerAviso(t *testing.T, c *websocket.Conn, quien string) negocio.Aviso {
	t.Helper()
	ctx, cancelar := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelar()
	var av negocio.Aviso
	if err := wsjson.Read(ctx, c, &av); err != nil {
		t.Fatalf("a %s no le llegó el aviso: %v", quien, err)
	}
	return av
}

func TestRetoAvisaALosQueAdministran(t *testing.T) {
	wsURL, httpURL, _ := servidorDelReto(t)
	conns := conectarATodos(t, wsURL)

	// ana se inscribe al evento 31 con su token, por la ruta de verdad.
	token, err := tokensDePrueba.Emitir(ana)
	if err != nil {
		t.Fatal(err)
	}
	pedido, _ := http.NewRequest("POST", httpURL+"/eventos/31/inscripciones", nil)
	pedido.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(pedido)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("se esperaba 201 al inscribirse y llegó %d", resp.StatusCode)
	}

	// Los dos organizadores conectados lo reciben, con tipo, id y evento.
	for _, quien := range []string{"luis", "marta"} {
		av := leerAviso(t, conns[quien], quien)
		if av.Tipo != "inscripcion_nueva" || av.ID != 9 || av.EventoID != 31 {
			t.Fatalf("a %s le llegó %+v", quien, av)
		}
	}
	// Las jugadoras no: ni eva ni ana, que fue quien se inscribió.
	noLlega(t, conns["ana"], "ana")
	noLlega(t, conns["eva"], "eva")
}

func TestRetoElAvisoDelDuenoSigueSoloParaElDueno(t *testing.T) {
	wsURL, _, avisos := servidorDelReto(t)
	conns := conectarATodos(t, wsURL)

	avisos.Enviar(ana.ID, negocio.Aviso{Tipo: "estado_cambiado", ID: 9, Estado: "confirmada"})

	if av := leerAviso(t, conns["ana"], "ana"); av.Tipo != "estado_cambiado" || av.ID != 9 {
		t.Fatalf("a ana le llegó %+v", av)
	}
	for _, quien := range []string{"eva", "luis", "marta"} {
		noLlega(t, conns[quien], quien)
	}
}
