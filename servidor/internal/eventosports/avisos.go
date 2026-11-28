package eventosports

import (
	"context"
	"eventosports/internal/negocio"
	"eventosports/internal/sesion"
	"net/http"
	"sync"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

// Avisos guarda, por usuario, los canales de sus conexiones abiertas.
// La llave es el id del usuario: el subject del token, el mismo número que
// guarda inscripciones.jugador_id. Una persona puede tener varias
// conexiones (el celular y la computadora); otra persona del mismo rol
// tiene su propia lista y no ve las de nadie más.
type Avisos struct {
	tokens  *sesion.Tokens
	origen  string // el front que puede abrir el WebSocket (el mismo de CORS)
	mu      sync.Mutex
	canales map[uint][]chan negocio.Aviso
}

func NuevosAvisos(tokens *sesion.Tokens, origen string) *Avisos {
	return &Avisos{tokens: tokens, origen: origen, canales: map[uint][]chan negocio.Aviso{}}
}

func (a *Avisos) Suscribir(usuarioID uint) chan negocio.Aviso {
	c := make(chan negocio.Aviso, 8)
	a.mu.Lock()
	defer a.mu.Unlock()
	a.canales[usuarioID] = append(a.canales[usuarioID], c)
	return c
}

func (a *Avisos) Quitar(usuarioID uint, c chan negocio.Aviso) {
	a.mu.Lock()
	defer a.mu.Unlock()
	lista := a.canales[usuarioID]
	for i := range lista {
		if lista[i] == c {
			lista = append(lista[:i], lista[i+1:]...)
			break
		}
	}
	if len(lista) == 0 {
		delete(a.canales, usuarioID) // ya no tiene conexiones abiertas
		return
	}
	a.canales[usuarioID] = lista
}

// Enviar escribe solo en los canales de ese usuario. Cumple negocio.Avisador.
func (a *Avisos) Enviar(usuarioID uint, av negocio.Aviso) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, c := range a.canales[usuarioID] {
		select {
		case c <- av:
		default: // conexión lenta: se pierde este aviso, no se frena a los demás
		}
	}
}

// Conectar: GET /ws?token=… Va FUERA del guardia: el navegador no deja
// poner Authorization en un WebSocket, así que el token se revisa aquí,
// antes de Accept. Sin sesión: 401 y no se abre nada.
func (a *Avisos) Conectar(w http.ResponseWriter, r *http.Request) {
	s, err := a.tokens.Leer(r.URL.Query().Get("token"))
	if err != nil {
		responderError(w, http.StatusUnauthorized, "sin sesión: falta el token o no es válido")
		return
	}
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{a.origen}, // el mismo origen que CORS
	})
	if err != nil {
		return // Accept ya respondió el error (403 si el origen no es el permitido)
	}
	defer conn.CloseNow()
	canal := a.Suscribir(s.UsuarioID)
	defer a.Quitar(s.UsuarioID, canal)
	// context.Background(), no r.Context(): este WebSocket solo envía, y
	// CloseRead avisa con ctx cuando el cliente se va.
	ctx := conn.CloseRead(context.Background())
	for {
		select {
		case <-ctx.Done():
			return // cerró la pestaña
		case av := <-canal:
			if err := wsjson.Write(ctx, conn, av); err != nil {
				return
			}
		}
	}
}
