package datos

import (
	"eventosports/internal/modelo"
	"eventosports/internal/sesion"
	"gorm.io/gorm"
	"time"
)

// Migrar crea todas las entidades, incluso las que aun no tienen endpoints.
func Migrar(db *gorm.DB) error {
	if err := db.AutoMigrate(&sesion.Usuario{}, &modelo.Usuario{}, &modelo.Sede{}, &modelo.Evento{}, &modelo.Inscripcion{}, &modelo.Pago{}); err != nil {
		return err
	}
	// Una cancelacion permite otra reserva; dos inscripciones activas no.
	return db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS inscripcion_activa_unica ON inscripciones (evento_id, jugador_id) WHERE estado <> 'cancelada'").Error
}

// SembrarDemo solo se invoca con DEMO_SEED=true. No borra reservas ni pagos.
// Las cuentas y los perfiles de ejemplo comparten ID: el subject del JWT
// identifica el perfil del negocio, sin aceptar identificadores del cliente.
func SembrarDemo(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := sesion.Sembrar(tx); err != nil {
			return err
		}
		for _, nombre := range []string{"ana", "luis"} {
			var cuenta sesion.Usuario
			if err := tx.Where("nombre = ?", nombre).First(&cuenta).Error; err != nil {
				return err
			}
			perfil := modelo.Usuario{ID: cuenta.ID, Nombre: cuenta.Nombre, Correo: nombre + "@eventosports.example", Rol: cuenta.Rol}
			if err := tx.FirstOrCreate(&perfil, modelo.Usuario{ID: cuenta.ID}).Error; err != nil {
				return err
			}
		}
		var organizador sesion.Usuario
		if err := tx.Where("nombre = ?", "luis").First(&organizador).Error; err != nil {
			return err
		}
		sede := modelo.Sede{ID: 3, Nombre: "Cancha El Palmar", Ciudad: "Manta", Referencia: "Junto al parque"}
		if err := tx.FirstOrCreate(&sede, modelo.Sede{ID: 3}).Error; err != nil {
			return err
		}
		for _, e := range []modelo.Evento{
			{ID: 25, Titulo: "Futbol de demostracion", Deporte: "futbol", Nivel: "casual", Inicio: time.Now().AddDate(0, 0, 30), Cupos: 14, Precio: 2.50, SedeID: 3, OrganizadorID: organizador.ID},
			{ID: 26, Titulo: "Padel gratuito de demostracion", Deporte: "padel", Nivel: "amateur", Inicio: time.Now().AddDate(0, 0, 31), Cupos: 8, Precio: 0, SedeID: 3, OrganizadorID: organizador.ID},
		} {
			if err := tx.FirstOrCreate(&e, modelo.Evento{ID: e.ID}).Error; err != nil {
				return err
			}
		}
		// Los ejemplos usan IDs fijos; el siguiente INSERT debe continuar despues del maximo.
		for _, tabla := range []string{"usuarios", "sedes", "eventos"} {
			if err := tx.Exec("SELECT setval(pg_get_serial_sequence('" + tabla + "', 'id'), COALESCE((SELECT MAX(id) FROM " + tabla + "), 1))").Error; err != nil {
				return err
			}
		}
		return nil
	})
}
