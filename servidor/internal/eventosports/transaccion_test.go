package eventosports

import (
	"context"
	"strings"
	"testing"

	"eventosports/internal/datos"
	"eventosports/internal/modelo"
	"eventosports/internal/negocio"
)

func TestPagoYEstadoSeReviertenSiFallaElEstado(t *testing.T) {
	for _, adaptador := range []string{"GORM", "sqlc"} {
		t.Run(adaptador, func(t *testing.T) {
			db := baseDePrueba(t)
			i := modelo.Inscripcion{EventoID: 25, JugadorID: 1, Estado: "reservada"}
			if err := db.Create(&i).Error; err != nil {
				t.Fatal(err)
			}
			// El pago se inserta primero. Esta restricción hace fallar la segunda escritura.
			if err := db.Exec("ALTER TABLE inscripciones ADD CONSTRAINT rechazar_estado CHECK (estado <> 'pago_por_verificar')").Error; err != nil {
				t.Fatal(err)
			}

			var almacen negocio.Almacen = &datos.AlmacenGORM{DB: db}
			if adaptador == "sqlc" {
				almacen = &datos.AlmacenSQLC{AlmacenGORM: &datos.AlmacenGORM{DB: db}}
			}
			servicio := &negocio.Servicio{Almacen: almacen}
			_, err := servicio.CambiarEstado(context.Background(), i.ID,
				negocio.CambioEstado{Estado: "pago_por_verificar", Banco: "Banco", Referencia: "FALLO-FORZADO"},
				negocio.Actor{ID: 1, Rol: "jugador"})
			if err == nil || !strings.Contains(err.Error(), "rechazar_estado") {
				t.Fatalf("se esperaba un fallo en la segunda escritura: %v", err)
			}

			var pagos int64
			if err := db.Model(&modelo.Pago{}).Where("inscripcion_id = ?", i.ID).Count(&pagos).Error; err != nil {
				t.Fatal(err)
			}
			if err := db.First(&i, i.ID).Error; err != nil {
				t.Fatal(err)
			}
			if pagos != 0 || i.Estado != "reservada" {
				t.Fatalf("transacción incompleta: pagos=%d, estado=%s", pagos, i.Estado)
			}
		})
	}
}
