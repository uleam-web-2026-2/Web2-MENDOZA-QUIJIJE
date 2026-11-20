package datos

import (
	"errors"

	"gorm.io/gorm"
)

// Semana 12 · reto. Lo único que puede hacer el usuario de la API, tabla por
// tabla. Ninguna tiene DELETE (la API no borra registros) ni TRUNCATE.
var permisosAPI = []struct{ tabla, permisos string }{
	{"cuentas", "SELECT"},  // POST /login lee nombre, rol y resumen de la clave
	{"usuarios", "SELECT"}, // perfiles del organizador y del jugador
	{"sedes", "SELECT"},    // filtro ?ciudad= y vista del evento
	// El evento se crea, pero nunca se cambia. UPDATE (cupos) está solo
	// porque PostgreSQL pide UPDATE en alguna columna para SELECT ... FOR
	// UPDATE, el bloqueo que evita vender dos veces el último cupo.
	{"eventos", "SELECT, INSERT, UPDATE (cupos)"},
	{"inscripciones", "SELECT, INSERT, UPDATE (estado)"}, // solo cambia el estado
	{"pagos", "SELECT, INSERT, UPDATE (verificado)"},     // solo se marca verificado
}

// Las tablas en las que la API inserta toman el id de una secuencia, que es
// otro objeto de la base y tiene su propio permiso.
var secuenciasAPI = []string{"eventos_id_seq", "inscripciones_id_seq", "pagos_id_seq"}

// PrepararUsuarioAPI crea (o actualiza) el usuario de la API sin
// superusuario y le deja exactamente los permisos de permisosAPI. La corre
// el dueño de las tablas en go run . -migrar, después de Migrar.
func PrepararUsuarioAPI(db *gorm.DB, usuario, clave string) error {
	if usuario == "" || clave == "" {
		return errors.New("falta el usuario o la clave de la API")
	}
	return db.Transaction(func(tx *gorm.DB) error {
		var existe bool
		if err := tx.Raw("SELECT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = ?)", usuario).Scan(&existe).Error; err != nil {
			return err
		}
		crear := "SELECT format('CREATE ROLE %I WITH LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS PASSWORD %L', ?::text, ?::text)"
		if existe {
			crear = "SELECT format('ALTER ROLE %I WITH LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS PASSWORD %L', ?::text, ?::text)"
		}
		if err := ordenConFormato(tx, crear, usuario, clave); err != nil {
			return err
		}
		// Se quita todo y se vuelve a dar solo lo de la lista: si alguien
		// agregó un permiso a mano, la próxima migración lo borra.
		for _, consulta := range []string{
			"SELECT format('GRANT CONNECT ON DATABASE %I TO %I', current_database(), ?::text)",
			"SELECT format('GRANT USAGE ON SCHEMA public TO %I', ?::text)",
			"SELECT format('REVOKE ALL ON ALL TABLES IN SCHEMA public FROM %I', ?::text)",
			"SELECT format('REVOKE ALL ON ALL SEQUENCES IN SCHEMA public FROM %I', ?::text)",
		} {
			if err := ordenConFormato(tx, consulta, usuario); err != nil {
				return err
			}
		}
		for _, p := range permisosAPI {
			if err := ordenConFormato(tx, "SELECT format('GRANT %s ON TABLE %I TO %I', ?::text, ?::text, ?::text)", p.permisos, p.tabla, usuario); err != nil {
				return err
			}
		}
		for _, s := range secuenciasAPI {
			if err := ordenConFormato(tx, "SELECT format('GRANT USAGE ON SEQUENCE %I TO %I', ?::text, ?::text)", s, usuario); err != nil {
				return err
			}
		}
		return nil
	})
}

// ordenConFormato arma la orden en PostgreSQL con format(): %I cita un
// nombre y %L un valor, y los valores viajan como parámetros. Hace falta
// porque CREATE ROLE, ALTER ROLE y GRANT no aceptan parámetros ($1). La
// plantilla siempre es un texto fijo del código.
func ordenConFormato(tx *gorm.DB, plantilla string, valores ...any) error {
	var orden string
	if err := tx.Raw(plantilla, valores...).Scan(&orden).Error; err != nil {
		return err
	}
	return tx.Exec(orden).Error
}
