package eventosports

import (
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
)

// Semana 12 · punto 1. El filtro ?ciudad= es el único texto libre del
// cliente que llega al SQL. Va como parámetro (?), así que PostgreSQL busca
// una ciudad que se llame literalmente "' OR 1=1 --": no hay ninguna. Si el
// texto se pegara en el SQL, la condición quedaría siempre verdadera y
// llegarían todos los eventos. Necesita la base, como las demás pruebas de
// integración (TEST_DATABASE_URL o servidor/.env).
func TestFiltroConComillaNoDevuelveNada(t *testing.T) {
	h := servidorDB(baseDePrueba(t))
	ataque := url.QueryEscape("' OR 1=1 --")

	normal := eventosDe(t, h, "/eventos?ciudad=Manta")
	if len(normal) == 0 {
		t.Fatal("con ?ciudad=Manta no llegó ningún evento; la siembra debía crear dos")
	}
	if inyectado := eventosDe(t, h, "/eventos?ciudad="+ataque); len(inyectado) != 0 {
		t.Fatalf("con ' OR 1=1 -- llegaron %d eventos; se esperaban 0", len(inyectado))
	}

	// deporte y nivel solo aceptan valores de una lista: el ataque ni llega a la base.
	esperar(t, llamar(h, "GET", "/eventos?deporte="+ataque, "", ""), 422)

	// El {id} se convierte a número antes de llegar a la base.
	token := login(t, h, "ana", "ana-2026")
	esperar(t, llamar(h, "GET", "/inscripciones/1%20OR%201=1", "", token), 400)
}

func eventosDe(t *testing.T, h http.Handler, ruta string) []map[string]any {
	t.Helper()
	rec := llamar(h, "GET", ruta, "", "")
	esperar(t, rec, 200)
	var lista []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &lista); err != nil {
		t.Fatalf("%s no devolvió una lista JSON: %v", ruta, err)
	}
	return lista
}
