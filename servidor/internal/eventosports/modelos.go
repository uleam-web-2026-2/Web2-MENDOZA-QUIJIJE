package eventosports

import "time"

// Structs del hito 1. Las relaciones crean claves foraneas al migrar PostgreSQL.

type Usuario struct {
	ID       uint    `gorm:"primaryKey" json:"id"`
	Nombre   string  `gorm:"not null" json:"nombre"`
	Correo   string  `gorm:"not null;uniqueIndex" json:"correo"`
	Rol      string  `gorm:"not null" json:"rol"` // jugador | organizador
	RUC      *string `json:"ruc,omitempty"`
	WhatsApp *string `json:"whatsapp,omitempty"`
}

type Sede struct {
	ID         uint     `gorm:"primaryKey" json:"id"`
	Nombre     string   `gorm:"not null" json:"nombre"`
	Ciudad     string   `gorm:"not null" json:"ciudad"`
	Referencia string   `gorm:"not null" json:"referencia"`
	Latitud    *float64 `json:"latitud,omitempty"`
	Longitud   *float64 `json:"longitud,omitempty"`
}

type Evento struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Titulo        string    `gorm:"not null" json:"titulo"`
	Deporte       string    `gorm:"not null" json:"deporte"` // futbol | padel | tenis | basquet | natacion
	Nivel         string    `gorm:"not null" json:"nivel"`   // casual | amateur | competitivo
	Inicio        time.Time `gorm:"not null" json:"inicio"`
	Cupos         int       `gorm:"not null" json:"cupos"`
	Precio        float64   `gorm:"type:numeric(8,2);not null;default:0" json:"precio"`
	OrganizadorID uint      `gorm:"not null" json:"organizador_id"`
	SedeID        uint      `gorm:"not null" json:"sede_id"`
	Organizador   Usuario   `gorm:"foreignKey:OrganizadorID" json:"-"`
	Sede          Sede      `gorm:"foreignKey:SedeID" json:"-"`
}

type Inscripcion struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	EventoID  uint      `gorm:"not null;index" json:"evento_id"`
	JugadorID uint      `gorm:"not null;index" json:"jugador_id"`
	Estado    string    `gorm:"not null;default:reservada" json:"estado"` // reservada | pago_por_verificar | confirmada | cancelada
	Creado    time.Time `gorm:"autoCreateTime" json:"creado"`
	Evento    Evento    `gorm:"foreignKey:EventoID" json:"-"`
	Jugador   Usuario   `gorm:"foreignKey:JugadorID" json:"-"`
}

type Pago struct {
	ID            uint        `gorm:"primaryKey" json:"id"`
	InscripcionID uint        `gorm:"not null" json:"inscripcion_id"`
	Monto         float64     `gorm:"type:numeric(8,2);not null" json:"monto"`
	Banco         string      `gorm:"not null" json:"banco"`
	Referencia    string      `gorm:"not null" json:"referencia"`
	Verificado    bool        `gorm:"not null;default:false" json:"verificado"`
	Creado        time.Time   `gorm:"autoCreateTime" json:"creado"`
	Inscripcion   Inscripcion `gorm:"foreignKey:InscripcionID" json:"-"`
}

func (Usuario) TableName() string     { return "usuarios" }
func (Sede) TableName() string        { return "sedes" }
func (Evento) TableName() string      { return "eventos" }
func (Inscripcion) TableName() string { return "inscripciones" }
func (Pago) TableName() string        { return "pagos" }
