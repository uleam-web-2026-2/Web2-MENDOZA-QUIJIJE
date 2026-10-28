package negocio

import (
	"context"
	"errors"
	"eventosports/internal/modelo"
	"strings"
)

func vista(q Consulta, e modelo.Evento) (EventoVista, error) {
	sede, err := q.Sede(e.SedeID)
	if err != nil {
		return EventoVista{}, err
	}
	n, err := q.Confirmadas(e.ID)
	if err != nil {
		return EventoVista{}, err
	}
	return EventoVista{Evento: e, Sede: sede, CuposLibres: e.Cupos - int(n)}, nil
}
func (s *Servicio) ListarEventos(ctx context.Context, f Filtro) ([]EventoVista, error) {
	if (f.Deporte != "" && !deportes[f.Deporte]) || (f.Nivel != "" && !niveles[f.Nivel]) {
		return nil, rechazar(Regla, "filtro fuera de la lista")
	}
	q := s.Almacen.ConContexto(ctx)
	eventos, err := q.Eventos(f)
	if err != nil {
		return nil, err
	}
	lista := []EventoVista{}
	for _, e := range eventos {
		v, err := vista(q, e)
		if err != nil {
			return nil, err
		}
		lista = append(lista, v)
	}
	return lista, nil
}
func (s *Servicio) VerEvento(ctx context.Context, id uint, a Actor) (EventoVista, error) {
	q := s.Almacen.ConContexto(ctx)
	e, err := q.Evento(id)
	if err != nil {
		return EventoVista{}, err
	}
	v, err := vista(q, e)
	if err != nil {
		return v, err
	}
	if a.Rol == "jugador" {
		i, err := q.UltimaInscripcion(id, a.ID)
		if err == nil {
			v.MiInscripcion = &i
		} else if !errors.Is(err, ErrNoExiste) {
			return v, err
		}
	}
	return v, nil
}
func (s *Servicio) CrearEvento(ctx context.Context, c CrearEvento, a Actor) (EventoVista, error) {
	if strings.TrimSpace(c.Titulo) == "" || !deportes[c.Deporte] || !niveles[c.Nivel] || c.Cupos <= 0 || c.Precio < 0 || !c.Inicio.After(s.reloj()) {
		return EventoVista{}, rechazar(Regla, "datos del evento invalidos: revise titulo, deporte, nivel, cupos, precio e inicio futuro")
	}
	if a.Rol != "organizador" {
		return EventoVista{}, rechazar(Prohibido, "solo un organizador puede publicar eventos")
	}
	q := s.Almacen.ConContexto(ctx)
	_, err := q.Sede(c.SedeID)
	if errors.Is(err, ErrNoExiste) {
		return EventoVista{}, rechazar(Regla, "la sede no existe")
	}
	if err != nil {
		return EventoVista{}, err
	}
	e := modelo.Evento{Titulo: c.Titulo, Deporte: c.Deporte, Nivel: c.Nivel, Inicio: c.Inicio, Cupos: c.Cupos, Precio: c.Precio, SedeID: c.SedeID, OrganizadorID: a.ID}
	if err := q.CrearEvento(&e); err != nil {
		return EventoVista{}, err
	}
	return vista(q, e)
}
func (s *Servicio) CrearInscripcion(ctx context.Context, id uint, a Actor) (modelo.Inscripcion, error) {
	if a.Rol != "jugador" {
		return modelo.Inscripcion{}, rechazar(Prohibido, "solo un jugador puede inscribirse")
	}
	var nueva modelo.Inscripcion
	err := s.Almacen.Transaccion(ctx, func(tx Operacion) error {
		e, err := tx.EventoBloqueado(id)
		if err != nil {
			return err
		}
		if !e.Inicio.After(s.reloj()) {
			return rechazar(Conflicto, "el evento ya empezo")
		}
		n, err := tx.Activas(id, a.ID)
		if err != nil {
			return err
		}
		if n > 0 {
			return rechazar(Conflicto, "ya tienes una inscripcion activa")
		}
		confirmadas, err := tx.Confirmadas(id)
		if err != nil {
			return err
		}
		if int64(e.Cupos) <= confirmadas {
			return rechazar(Conflicto, "no quedan cupos")
		}
		nueva = modelo.Inscripcion{EventoID: id, JugadorID: a.ID, Estado: "reservada", Creado: s.reloj()}
		return tx.CrearInscripcion(&nueva)
	})
	return nueva, err
}
func (s *Servicio) VerInscripcion(ctx context.Context, id uint, a Actor) (modelo.Inscripcion, error) {
	q := s.Almacen.ConContexto(ctx)
	i, err := q.Inscripcion(id)
	if err != nil {
		return i, err
	}
	e, err := q.Evento(i.EventoID)
	if err != nil {
		return i, err
	}
	if !(a.Rol == "jugador" && a.ID == i.JugadorID) && !(a.Rol == "organizador" && a.ID == e.OrganizadorID) {
		return i, rechazar(Prohibido, "esta inscripcion no es tuya ni de tus eventos")
	}
	return i, nil
}
func (s *Servicio) ListarInscripciones(ctx context.Context, id uint, a Actor) ([]InscripcionVista, error) {
	q := s.Almacen.ConContexto(ctx)
	e, err := q.Evento(id)
	if err != nil {
		return nil, err
	}
	if a.Rol != "organizador" || e.OrganizadorID != a.ID {
		return nil, rechazar(Prohibido, "el evento pertenece a otro organizador")
	}
	inscripciones, err := q.Inscripciones(id)
	if err != nil {
		return nil, err
	}
	lista := []InscripcionVista{}
	for _, i := range inscripciones {
		u, err := q.Usuario(i.JugadorID)
		if err != nil {
			return nil, err
		}
		v := InscripcionVista{Inscripcion: i, Jugador: JugadorVista{u.ID, u.Nombre}}
		p, err := q.UltimoPago(i.ID)
		if err == nil {
			v.UltimoPago = &p
		} else if !errors.Is(err, ErrNoExiste) {
			return nil, err
		}
		lista = append(lista, v)
	}
	return lista, nil
}
func (s *Servicio) CambiarEstado(ctx context.Context, id uint, c CambioEstado, a Actor) (modelo.Inscripcion, error) {
	if !estados[c.Estado] {
		return modelo.Inscripcion{}, rechazar(Regla, "estado fuera de la lista")
	}
	var i modelo.Inscripcion
	err := s.Almacen.Transaccion(ctx, func(tx Operacion) error {
		var err error
		i, err = tx.Inscripcion(id)
		if err != nil {
			return err
		}
		// Todos los cambios de cupos bloquean el evento antes de la inscripcion.
		e, err := tx.EventoBloqueado(i.EventoID)
		if err != nil {
			return err
		}
		i, err = tx.InscripcionBloqueada(id)
		if err != nil {
			return err
		}
		n, err := tx.Confirmadas(e.ID)
		if err != nil {
			return err
		}
		ahora := s.reloj()
		if err := ValidarCambio(i, e, a, c, n, ahora); err != nil {
			return err
		}
		if i.Estado == "reservada" && c.Estado == "pago_por_verificar" {
			p := modelo.Pago{InscripcionID: i.ID, Monto: e.Precio, Banco: c.Banco, Referencia: c.Referencia, Creado: ahora}
			if err := tx.CrearPago(&p); err != nil {
				return err
			}
		}
		if i.Estado == "pago_por_verificar" && c.Estado == "confirmada" {
			p, err := tx.UltimoPago(i.ID)
			if err != nil {
				return err
			}
			if err := tx.VerificarPago(p.ID); err != nil {
				return err
			}
		}
		i.Estado = c.Estado
		return tx.GuardarEstado(i)
	})
	return i, err
}
