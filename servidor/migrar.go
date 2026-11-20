package main

import (
	"errors"
	"eventosports/internal/configuracion"
	"eventosports/internal/datos"
	"net/url"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// migrarConDueno corre con go run . -migrar. Se conecta con el dueño de las
// tablas (MIGRACION_DATABASE_URL), crea o cambia las tablas, siembra los
// ejemplos si DEMO_SEED=true y deja al usuario de DATABASE_URL con los
// permisos justos. La API, al arrancar normal, ya no migra.
func migrarConDueno(c configuracion.Config, cfg *gorm.Config) error {
	if c.Migracion == "" {
		return errors.New("falta MIGRACION_DATABASE_URL (el usuario dueño de las tablas)")
	}
	api, err := url.Parse(c.Base)
	if err != nil || api.User == nil || api.User.Username() == "" {
		return errors.New("DATABASE_URL debe traer el usuario de la API")
	}
	dueno, err := url.Parse(c.Migracion)
	if err != nil || dueno.User == nil {
		return errors.New("MIGRACION_DATABASE_URL debe traer el usuario dueño de las tablas")
	}
	if api.User.Username() == dueno.User.Username() {
		return errors.New("la API no puede conectarse con el dueño de las tablas: use otro usuario en DATABASE_URL")
	}
	clave, _ := api.User.Password()

	db, err := gorm.Open(postgres.Open(c.Migracion), cfg)
	if err != nil {
		return errors.New("no se pudo conectar con MIGRACION_DATABASE_URL")
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()
	if err := datos.Migrar(db); err != nil {
		return err
	}
	if c.Demo {
		if err := datos.SembrarDemo(db); err != nil {
			return err
		}
	}
	return datos.PrepararUsuarioAPI(db, api.User.Username(), clave)
}
