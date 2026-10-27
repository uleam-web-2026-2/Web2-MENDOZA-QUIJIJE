package eventosports

import (
	"encoding/json"
	"errors"
	"eventosports/internal/sesion"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Manejador struct {
	DB    *gorm.DB
	ahora func() time.Time
}

func (m *Manejador) reloj() time.Time {
	if m.ahora != nil {
		return m.ahora()
	}
	return time.Now()
}
func responder(w http.ResponseWriter, codigo int, cuerpo any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(codigo)
	_ = json.NewEncoder(w).Encode(cuerpo)
}
func responderError(w http.ResponseWriter, codigo int, mensaje string) {
	responder(w, codigo, map[string]string{"error": mensaje})
}
func idRuta(w http.ResponseWriter, r *http.Request) (uint, bool) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil || id == 0 {
		responderError(w, 400, "el id debe ser un entero positivo")
		return 0, false
	}
	return uint(id), true
}
func leerJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		responderError(w, 400, "JSON mal formado")
		return false
	}
	return true
}
func falloDB(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		responderError(w, 404, "el recurso no existe")
	} else {
		responderError(w, 500, "no se pudo acceder a los datos")
	}
	return true
}

type eventoVista struct {
	Evento
	Sede          Sede         `json:"sede"`
	CuposLibres   int          `json:"cupos_libres"`
	MiInscripcion *Inscripcion `json:"mi_inscripcion,omitempty"`
}

func vista(db *gorm.DB, e Evento) (eventoVista, error) {
	var sede Sede
	if err := db.First(&sede, e.SedeID).Error; err != nil {
		return eventoVista{}, err
	}
	var usados int64
	if err := db.Model(&Inscripcion{}).Where("evento_id = ? AND estado = ?", e.ID, "confirmada").Count(&usados).Error; err != nil {
		return eventoVista{}, err
	}
	return eventoVista{Evento: e, Sede: sede, CuposLibres: e.Cupos - int(usados)}, nil
}
func (m *Manejador) listarEventos(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if (q.Get("deporte") != "" && !deportes[q.Get("deporte")]) || (q.Get("nivel") != "" && !niveles[q.Get("nivel")]) {
		responderError(w, 422, "filtro fuera de la lista")
		return
	}
	db := m.DB.WithContext(r.Context()).Model(&Evento{})
	for _, campo := range []string{"deporte", "nivel"} {
		if valor := q.Get(campo); valor != "" {
			db = db.Where(campo+" = ?", valor)
		}
	}
	if ciudad := q.Get("ciudad"); ciudad != "" {
		db = db.Where("sede_id IN (?)", m.DB.Model(&Sede{}).Select("id").Where("LOWER(ciudad) = LOWER(?)", ciudad))
	}
	var eventos []Evento
	if falloDB(w, db.Order("id").Find(&eventos).Error) {
		return
	}
	lista := []eventoVista{}
	for _, e := range eventos {
		v, err := vista(m.DB.WithContext(r.Context()), e)
		if falloDB(w, err) {
			return
		}
		lista = append(lista, v)
	}
	responder(w, 200, lista)
}
func (m *Manejador) verEvento(w http.ResponseWriter, r *http.Request) {
	id, ok := idRuta(w, r)
	if !ok {
		return
	}
	var e Evento
	db := m.DB.WithContext(r.Context())
	if falloDB(w, db.First(&e, id).Error) {
		return
	}
	v, err := vista(db, e)
	if err == nil {
		if s, ok := sesion.De(r); ok && s.Rol == sesion.RolJugador {
			var i Inscripcion
			err = db.Where("evento_id = ? AND jugador_id = ?", e.ID, s.UsuarioID).Order("id DESC").First(&i).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err = nil
			} else if err == nil {
				v.MiInscripcion = &i
			}
		}
	}
	if !falloDB(w, err) {
		responder(w, 200, v)
	}
}

type cuerpoEvento struct {
	Titulo  string    `json:"titulo"`
	Deporte string    `json:"deporte"`
	Nivel   string    `json:"nivel"`
	Inicio  time.Time `json:"inicio"`
	Cupos   int       `json:"cupos"`
	Precio  float64   `json:"precio"`
	SedeID  uint      `json:"sede_id"`
}

func (m *Manejador) crearEvento(w http.ResponseWriter, r *http.Request) {
	var c cuerpoEvento
	if !leerJSON(w, r, &c) {
		return
	}
	if strings.TrimSpace(c.Titulo) == "" || !deportes[c.Deporte] || !niveles[c.Nivel] || c.Cupos <= 0 || c.Precio < 0 || !c.Inicio.After(m.reloj()) {
		responderError(w, 422, "datos del evento invalidos: revise titulo, deporte, nivel, cupos, precio e inicio futuro")
		return
	}
	db := m.DB.WithContext(r.Context())
	var sede Sede
	err := db.First(&sede, c.SedeID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		responderError(w, 422, "la sede no existe")
		return
	}
	if falloDB(w, err) {
		return
	}
	s, _ := sesion.De(r)
	e := Evento{Titulo: c.Titulo, Deporte: c.Deporte, Nivel: c.Nivel, Inicio: c.Inicio, Cupos: c.Cupos, Precio: c.Precio, SedeID: c.SedeID, OrganizadorID: s.UsuarioID}
	if falloDB(w, db.Create(&e).Error) {
		return
	}
	v, err := vista(db, e)
	if !falloDB(w, err) {
		responder(w, 201, v)
	}
}

type errorNegocio struct {
	codigo  int
	mensaje string
}

func (e *errorNegocio) Error() string { return e.mensaje }
func rechazo(c int, s string) error   { return &errorNegocio{c, s} }
func resultado(w http.ResponseWriter, err error) bool {
	var e *errorNegocio
	if errors.As(err, &e) {
		responderError(w, e.codigo, e.mensaje)
		return true
	}
	return falloDB(w, err)
}
func (m *Manejador) crearInscripcion(w http.ResponseWriter, r *http.Request) {
	id, ok := idRuta(w, r)
	if !ok {
		return
	}
	s, _ := sesion.De(r)
	var nueva Inscripcion
	err := m.DB.WithContext(r.Context()).Transaction(func(tx *gorm.DB) error {
		var e Evento
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&e, id).Error; err != nil {
			return err
		}
		if !e.Inicio.After(m.reloj()) {
			return rechazo(409, "el evento ya empezo")
		}
		var n int64
		if err := tx.Model(&Inscripcion{}).Where("evento_id = ? AND jugador_id = ? AND estado <> ?", id, s.UsuarioID, "cancelada").Count(&n).Error; err != nil {
			return err
		}
		if n > 0 {
			return rechazo(409, "ya tienes una inscripcion activa")
		}
		v, err := vista(tx, e)
		if err != nil {
			return err
		}
		if v.CuposLibres <= 0 {
			return rechazo(409, "no quedan cupos")
		}
		nueva = Inscripcion{EventoID: id, JugadorID: s.UsuarioID, Estado: "reservada", Creado: m.reloj()}
		return tx.Create(&nueva).Error
	})
	if !resultado(w, err) {
		responder(w, 201, nueva)
	}
}

type jugadorVista struct {
	ID     uint   `json:"id"`
	Nombre string `json:"nombre"`
}
type inscripcionVista struct {
	Inscripcion
	Jugador    jugadorVista `json:"jugador"`
	UltimoPago *Pago        `json:"ultimo_pago,omitempty"`
}

func (m *Manejador) listarInscripciones(w http.ResponseWriter, r *http.Request) {
	id, ok := idRuta(w, r)
	if !ok {
		return
	}
	s, _ := sesion.De(r)
	db := m.DB.WithContext(r.Context())
	var e Evento
	if falloDB(w, db.First(&e, id).Error) {
		return
	}
	if e.OrganizadorID != s.UsuarioID {
		responderError(w, 403, "el evento pertenece a otro organizador")
		return
	}
	var inscripciones []Inscripcion
	if falloDB(w, db.Where("evento_id = ?", id).Order("id").Find(&inscripciones).Error) {
		return
	}
	lista := []inscripcionVista{}
	for _, i := range inscripciones {
		var u Usuario
		if falloDB(w, db.First(&u, i.JugadorID).Error) {
			return
		}
		v := inscripcionVista{Inscripcion: i, Jugador: jugadorVista{u.ID, u.Nombre}}
		var p Pago
		err := db.Where("inscripcion_id = ?", i.ID).Order("id DESC").First(&p).Error
		if err == nil {
			v.UltimoPago = &p
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			falloDB(w, err)
			return
		}
		lista = append(lista, v)
	}
	responder(w, 200, lista)
}
