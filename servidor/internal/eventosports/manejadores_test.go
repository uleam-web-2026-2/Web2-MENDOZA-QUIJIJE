package eventosports

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

var organizador = map[string]string{"X-Usuario-ID": "7", "X-Rol": "organizador"}
var jugador = map[string]string{"X-Usuario-ID": "12", "X-Rol": "jugador"}

func llamar(h http.Handler, metodo, ruta, cuerpo string, cabeceras map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(metodo, ruta, strings.NewReader(cuerpo))
	for k, v := range cabeceras {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func esperar(t *testing.T, rec *httptest.ResponseRecorder, codigo int) {
	t.Helper()
	if rec.Code != codigo {
		t.Fatalf("código %d, se esperaba %d; cuerpo: %s", rec.Code, codigo, rec.Body.String())
	}
}

// 1. POST /eventos con un cuerpo mal formado responde 400, antes de tocar los datos.
func TestCrearEvento_JSONRoto(t *testing.T) {
	h := NuevoServidor().Rutas()
	esperar(t, llamar(h, "POST", "/eventos", `{"titulo":`, organizador), http.StatusBadRequest)
}

// 2. PATCH con un estado que no está en la lista cerrada responde 422.
func TestCambiarEstado_FueraDeLaLista(t *testing.T) {
	h := NuevoServidor().Rutas()
	esperar(t, llamar(h, "PATCH", "/inscripciones/118/estado", `{"estado":"vip"}`, jugador), http.StatusUnprocessableEntity)
}

// 3. La tabla de transiciones rechaza cancelada → confirmada, y el manejador responde 409.
func TestTransicion_CanceladaAConfirmada(t *testing.T) {
	if transicionPermitida("cancelada", "confirmada") {
		t.Fatal("cancelada → confirmada debería estar prohibida")
	}
	s := NuevoServidor()
	s.ahora = func() time.Time { return time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC) } // reloj fijo
	h := s.Rutas()
	rec := llamar(h, "POST", "/eventos/26/inscripciones", "", jugador) // el evento 26 es gratuito
	esperar(t, rec, http.StatusCreated)
	var i Inscripcion
	if err := json.Unmarshal(rec.Body.Bytes(), &i); err != nil {
		t.Fatal(err)
	}
	ruta := "/inscripciones/" + strconv.Itoa(int(i.ID)) + "/estado"
	esperar(t, llamar(h, "PATCH", ruta, `{"estado":"cancelada"}`, jugador), http.StatusOK)
	esperar(t, llamar(h, "PATCH", ruta, `{"estado":"confirmada"}`, jugador), http.StatusConflict)
}

// 4. POST /eventos con cupos 0 responde 422.
func TestCrearEvento_CuposCero(t *testing.T) {
	h := NuevoServidor().Rutas()
	cuerpo := `{"titulo":"Fútbol 7 mixto","deporte":"futbol","nivel":"casual","inicio":"2030-01-01T20:00:00-05:00","cupos":0,"precio":2.5,"sede_id":3}`
	esperar(t, llamar(h, "POST", "/eventos", cuerpo, organizador), http.StatusUnprocessableEntity)
}

// 5. GET /eventos/abc responde 400 (el id no es un número).
func TestVerEvento_IdNoNumerico(t *testing.T) {
	h := NuevoServidor().Rutas()
	esperar(t, llamar(h, "GET", "/eventos/abc", "", jugador), http.StatusBadRequest)
}

// 6. Un jugador no puede publicar eventos: POST /eventos responde 403.
func TestCrearEvento_RolJugador(t *testing.T) {
	h := NuevoServidor().Rutas()
	cuerpo := `{"titulo":"Fútbol 7 mixto","deporte":"futbol","nivel":"casual","inicio":"2030-01-01T20:00:00-05:00","cupos":14,"precio":2.5,"sede_id":3}`
	esperar(t, llamar(h, "POST", "/eventos", cuerpo, jugador), http.StatusForbidden)
}

// 7. GET /eventos entrega los dos eventos de ejemplo en orden estable.
func TestListarEventos_DosSemillas(t *testing.T) {
	h := NuevoServidor().Rutas()
	rec := llamar(h, "GET", "/eventos", "", nil)
	esperar(t, rec, http.StatusOK)
	var eventos []eventoVista
	if err := json.Unmarshal(rec.Body.Bytes(), &eventos); err != nil {
		t.Fatal(err)
	}
	if len(eventos) != 2 || eventos[0].ID != 25 || eventos[1].ID != 26 {
		t.Fatalf("se esperaban los eventos 25 y 26; llegaron: %+v", eventos)
	}
}

// 8. Un organizador puede publicar un evento con datos válidos.
func TestCrearEvento_OK(t *testing.T) {
	s := NuevoServidor()
	s.ahora = func() time.Time { return time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC) }
	h := s.Rutas()
	cuerpo := `{"titulo":"Fútbol 7 mixto","deporte":"futbol","nivel":"casual","inicio":"2026-10-15T20:00:00-05:00","cupos":14,"precio":2.5,"sede_id":3}`
	rec := llamar(h, "POST", "/eventos", cuerpo, organizador)
	esperar(t, rec, http.StatusCreated)
	var evento eventoVista
	if err := json.Unmarshal(rec.Body.Bytes(), &evento); err != nil {
		t.Fatal(err)
	}
	if evento.ID != 27 || evento.OrganizadorID != 7 || evento.CuposLibres != 14 {
		t.Fatalf("datos inesperados del evento creado: %+v", evento)
	}
}

// 9. El jugador registra una transferencia y el organizador puede verificarla.
func TestInscripcion_PagoYVerificacion(t *testing.T) {
	s := NuevoServidor()
	s.ahora = func() time.Time { return time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC) }
	h := s.Rutas()

	creada := llamar(h, "POST", "/eventos/25/inscripciones", "", jugador)
	esperar(t, creada, http.StatusCreated)
	var inscripcion Inscripcion
	if err := json.Unmarshal(creada.Body.Bytes(), &inscripcion); err != nil {
		t.Fatal(err)
	}
	if inscripcion.Estado != "reservada" {
		t.Fatalf("estado inicial inesperado: %s", inscripcion.Estado)
	}

	ruta := "/inscripciones/" + strconv.Itoa(int(inscripcion.ID)) + "/estado"
	pago := `{"estado":"pago_por_verificar","banco":"Banco Pichincha","referencia":"4839201"}`
	esperar(t, llamar(h, "PATCH", ruta, pago, jugador), http.StatusOK)

	verificada := llamar(h, "PATCH", ruta, `{"estado":"confirmada"}`, organizador)
	esperar(t, verificada, http.StatusOK)
	var final Inscripcion
	if err := json.Unmarshal(verificada.Body.Bytes(), &final); err != nil {
		t.Fatal(err)
	}
	if final.Estado != "confirmada" || !s.pagos[0].Verificado {
		t.Fatalf("no se verificó el pago y la inscripción: inscripción=%+v pago=%+v", final, s.pagos[0])
	}
}
