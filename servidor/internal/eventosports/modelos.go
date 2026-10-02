package eventosports

import "time"

// Structs de la sección 4 de la ficha. Las etiquetas de GORM quedan listas para cuando se conecte PostgreSQL;
// mientras tanto el servidor guarda los datos en memoria.

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
}

type Inscripcion struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	EventoID  uint      `gorm:"not null;index" json:"evento_id"`
	JugadorID uint      `gorm:"not null;index" json:"jugador_id"`
	Estado    string    `gorm:"not null;default:reservada" json:"estado"` // reservada | pago_por_verificar | confirmada | cancelada
	Creado    time.Time `gorm:"autoCreateTime" json:"creado"`
}

type Pago struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	InscripcionID uint      `gorm:"not null" json:"inscripcion_id"`
	Monto         float64   `gorm:"type:numeric(8,2);not null" json:"monto"`
	Banco         string    `gorm:"not null" json:"banco"`
	Referencia    string    `gorm:"not null" json:"referencia"`
	Verificado    bool      `gorm:"not null;default:false" json:"verificado"`
	Creado        time.Time `gorm:"autoCreateTime" json:"creado"`
}
