package configuracion

import (
	"errors"
	"github.com/joho/godotenv"
	"os"
)

type Config struct {
	Puerto, Base, Secreto string
	Demo                  bool
}

func Leer() (Config, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return Config{}, err
	}
	c := Config{Puerto: os.Getenv("PORT"), Base: os.Getenv("DATABASE_URL"), Secreto: os.Getenv("JWT_SECRET"), Demo: os.Getenv("DEMO_SEED") == "true"}
	if c.Puerto == "" {
		c.Puerto = "8080"
	}
	if c.Base == "" {
		return c, errors.New("falta DATABASE_URL")
	}
	if len(c.Secreto) < 32 {
		return c, errors.New("JWT_SECRET debe tener al menos 32 caracteres")
	}
	return c, nil
}
