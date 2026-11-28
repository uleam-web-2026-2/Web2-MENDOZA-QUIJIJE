package negocio

// Aviso es lo que se le cuenta a una persona cuando cambia algo suyo.
// Sale como JSON por el WebSocket: el hito 3 pide al menos tipo e id.
// En EventoSports el registro es la inscripción: ID es el de la inscripción.
type Aviso struct {
	Tipo     string `json:"tipo"`
	ID       uint   `json:"id"`
	Estado   string `json:"estado"`
	EventoID uint   `json:"evento_id,omitempty"` // reto: a qué evento se inscribieron
}

// Avisador es lo que el servicio necesita para avisar. El servicio no
// sabe si el aviso sale por WebSocket, por SSE o por correo: solo dice
// a quién y qué.
//
// «A quién» es el id del usuario (el mismo número que el subject del
// token y que inscripciones.jugador_id), no su nombre: así el dueño que
// guardó la inscripción y la llave con la que se busca en el mapa de
// avisos son el mismo dato, y el compilador no deja mezclarlos.
type Avisador interface {
	Enviar(usuarioID uint, av Aviso)
	// AvisarAlNegocio cuenta algo nuevo a quienes administran (reto del
	// taller 13). El servicio no nombra ese rol: lo decide quien implementa
	// el Avisador, con lo que le da main.go.
	AvisarAlNegocio(av Aviso)
}
