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

// conexion es una conexión abierta: su canal y el rol con el que entró,
// leído del token al conectar (reto del taller 13).
type conexion struct {
	rol   string
	canal chan negocio.Aviso
}

// Avisos guarda, por usuario, sus conexiones abiertas.
// La llave es el id del usuario: el subject del token, el mismo número que
// guarda inscripciones.jugador_id. Una persona puede tener varias
// conexiones (el celular y la computadora); otra persona del mismo rol
// tiene su propia lista y no ve las de nadie más.
type Avisos struct {
	tokens     *sesion.Tokens
	origen     string // el front que puede abrir el WebSocket (el mismo de CORS)
	rolNegocio string // el rol que administra: recibe AvisarAlNegocio (lo da main.go)
	mu         sync.Mutex
	conexiones map[uint][]conexion
}

func NuevosAvisos(tokens *sesion.Tokens, origen, rolNegocio string) *Avisos {
	return &Avisos{tokens: tokens, origen: origen, rolNegocio: rolNegocio, conexiones: map[uint][]conexion{}}
}

func (a *Avisos) Suscribir(usuarioID uint, rol string) chan negocio.Aviso {
	c := make(chan negocio.Aviso, 8)
	a.mu.Lock()
	defer a.mu.Unlock()
	a.conexiones[usuarioID] = append(a.conexiones[usuarioID], conexion{rol: rol, canal: c})
	return c
}

func (a *Avisos) Quitar(usuarioID uint, c chan negocio.Aviso) {
	a.mu.Lock()
	defer a.mu.Unlock()
	lista := a.conexiones[usuarioID]
	for i := range lista {
		if lista[i].canal == c {
			lista = append(lista[:i], lista[i+1:]...)
			break
		}
	}
	if len(lista) == 0 {
		delete(a.conexiones, usuarioID) // ya no tiene conexiones abiertas
		return
	}
	a.conexiones[usuarioID] = lista
}

// Enviar escribe solo en las conexiones de ese usuario. Cumple negocio.Avisador.
func (a *Avisos) Enviar(usuarioID uint, av negocio.Aviso) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, c := range a.conexiones[usuarioID] {
		mandar(c.canal, av)
	}
}

// AvisarAlNegocio escribe en todas las conexiones abiertas con el rol que
// administra, y en ninguna otra: el filtro es aquí, en el servidor, no en
// el cliente. Cumple negocio.Avisador.
func (a *Avisos) AvisarAlNegocio(av negocio.Aviso) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, lista := range a.conexiones {
		for _, c := range lista {
			if c.rol == a.rolNegocio {
				mandar(c.canal, av)
			}
		}
	}
}

// mandar no espera: si el canal está lleno (conexión lenta), se pierde
// este aviso y no se frena a los demás.
func mandar(c chan negocio.Aviso, av negocio.Aviso) {
	select {
	case c <- av:
	default:
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
	// El id y el rol salen del token firmado, no de algo que mande el cliente.
	canal := a.Suscribir(s.UsuarioID, s.Rol)
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
