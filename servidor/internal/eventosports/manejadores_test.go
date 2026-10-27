package eventosports

import (
	"encoding/base64"
	"encoding/json"
	"eventosports/internal/sesion"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func llamar(h http.Handler, metodo, ruta, cuerpo, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(metodo, ruta, strings.NewReader(cuerpo))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}
func esperar(t *testing.T, r *httptest.ResponseRecorder, c int) {
	t.Helper()
	if r.Code != c {
		t.Fatalf("se esperaba %d, llego %d: %s", c, r.Code, r.Body.String())
	}
}

func TestValidacionesAntesDeConsultarBase(t *testing.T) {
	for _, caso := range []struct {
		nombre, rol, metodo, ruta, cuerpo string
		codigo                            int
	}{
		{"JSON roto", sesion.RolOrganizador, "POST", "/eventos", `{"titulo":`, 400},
		{"cupos cero", sesion.RolOrganizador, "POST", "/eventos", `{"titulo":"Futbol","deporte":"futbol","nivel":"casual","inicio":"2099-01-01T20:00:00-05:00","cupos":0,"sede_id":3}`, 422},
		{"estado desconocido", sesion.RolJugador, "PATCH", "/inscripciones/1/estado", `{"estado":"vip"}`, 422},
		{"id no numerico", sesion.RolJugador, "GET", "/eventos/abc", "", 400},
		{"id cero", sesion.RolJugador, "GET", "/eventos/0", "", 400},
		{"filtro invalido", sesion.RolJugador, "GET", "/eventos?deporte=ajedrez", "", 422},
		{"rol no permitido", sesion.RolJugador, "POST", "/eventos", `{}`, 403},
		{"solo jugador", sesion.RolOrganizador, "POST", "/eventos/25/inscripciones", "", 403},
		{"solo organizador", sesion.RolJugador, "GET", "/eventos/25/inscripciones", "", 403},
	} {
		t.Run(caso.nombre, func(t *testing.T) {
			esperar(t, llamar(enrutadorCon(caso.rol), caso.metodo, caso.ruta, caso.cuerpo, ""), caso.codigo)
		})
	}
}
func TestCanceladaEsFinal(t *testing.T) {
	for estado := range estados {
		if transicionPermitida("cancelada", estado) {
			t.Fatal("cancelada no puede cambiar de estado")
		}
	}
}
func TestIdentidadNoSeAceptaDesdeCabeceras(t *testing.T) {
	req := httptest.NewRequest("POST", "/eventos", strings.NewReader(`{}`))
	req.Header.Set("X-Usuario-ID", "2")
	req.Header.Set("X-Rol", "organizador")
	rec := httptest.NewRecorder()
	enrutadorCon("").ServeHTTP(rec, req)
	esperar(t, rec, 401)
}
func TestTokensInvalidosNoLleganALaBase(t *testing.T) {
	secreto := "secreto-solo-de-las-pruebas"
	vencido, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "2", "rol": "organizador", "exp": time.Now().Add(-time.Hour).Unix()}).SignedString([]byte(secreto))
	sinExp, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "2", "rol": "organizador"}).SignedString([]byte(secreto))
	otro, _ := sesion.NuevosTokens("otro-secreto").Emitir(sesion.Usuario{ID: 2, Rol: sesion.RolOrganizador})
	token, _ := tokensDePrueba.Emitir(sesion.Usuario{ID: 1, Rol: sesion.RolJugador})
	partes := strings.Split(token, ".")
	payload, _ := base64.RawURLEncoding.DecodeString(partes[1])
	partes[1] = base64.RawURLEncoding.EncodeToString([]byte(strings.ReplaceAll(string(payload), "jugador", "organizador")))
	for nombre, valor := range map[string]string{"vencido": vencido, "sin vencimiento": sinExp, "otra firma": otro, "rol manipulado": strings.Join(partes, "."), "texto invalido": "basura"} {
		t.Run(nombre, func(t *testing.T) { esperar(t, llamar(enrutadorCon(""), "POST", "/eventos", `{}`, valor), 401) })
	}
}

// Cada prueba crea su propio esquema y lo elimina al finalizar. Nunca trunca
// ni borra tablas del esquema public ni registros del proyecto.
func baseDePrueba(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		valores, _ := godotenv.Read("../../.env")
		dsn = valores["DATABASE_URL"]
	}
	if dsn == "" {
		t.Skip("integracion PostgreSQL: configure TEST_DATABASE_URL o servidor/.env")
	}
	u, err := url.Parse(dsn)
	if err != nil || u.Scheme != "postgres" && u.Scheme != "postgresql" {
		t.Fatal("la URL de pruebas debe ser postgres:// o postgresql://")
	}
	cfg := &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)}
	admin, err := gorm.Open(postgres.Open(dsn), cfg)
	if err != nil {
		t.Fatal("no se pudo conectar a la base de pruebas")
	}
	sqlAdmin, _ := admin.DB()
	esquema := fmt.Sprintf("s7_test_%d", time.Now().UnixNano())
	if err := admin.Exec("CREATE SCHEMA " + esquema).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := admin.Exec("DROP SCHEMA " + esquema + " CASCADE").Error; err != nil {
			t.Error(err)
		}
		sqlAdmin.Close()
	})
	q := u.Query()
	q.Set("search_path", esquema)
	u.RawQuery = q.Encode()
	db, err := gorm.Open(postgres.Open(u.String()), cfg)
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	t.Cleanup(func() { sqlDB.Close() })
	if err := Migrar(db); err != nil {
		t.Fatal(err)
	}
	if err := SembrarDemo(db); err != nil {
		t.Fatal(err)
	}
	return db
}
func servidorDB(db *gorm.DB) http.Handler {
	r := chi.NewRouter()
	(&Manejador{DB: db}).Rutas(r, tokensDePrueba)
	return r
}
func login(t *testing.T, h http.Handler, usuario, clave string) string {
	t.Helper()
	rec := llamar(h, "POST", "/login", `{"usuario":"`+usuario+`","clave":"`+clave+`"}`, "")
	esperar(t, rec, 200)
	var v struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil || v.Token == "" {
		t.Fatal("login sin token")
	}
	return v.Token
}
func crearReserva(t *testing.T, h http.Handler, event uint, token string) Inscripcion {
	t.Helper()
	r := llamar(h, "POST", fmt.Sprintf("/eventos/%d/inscripciones", event), "", token)
	esperar(t, r, 201)
	var i Inscripcion
	if err := json.Unmarshal(r.Body.Bytes(), &i); err != nil {
		t.Fatal(err)
	}
	return i
}

func TestIntegracionSeisPeticionesYPersistencia(t *testing.T) {
	db := baseDePrueba(t)
	h := servidorDB(db)
	// La ruta protegida se elige del negocio: listar inscripciones del evento.
	esperar(t, llamar(h, "GET", "/eventos/25/inscripciones", "", ""), 401)
	org := login(t, h, "luis", "luis-2026")
	jug := login(t, h, "ana", "ana-2026")
	esperar(t, llamar(h, "GET", "/eventos/25/inscripciones", "", org), 200)
	i := crearReserva(t, h, 25, jug)
	detalle := llamar(h, "GET", "/eventos/25", "", jug)
	esperar(t, detalle, 200)
	var personal eventoVista
	if err := json.Unmarshal(detalle.Body.Bytes(), &personal); err != nil || personal.MiInscripcion == nil || personal.MiInscripcion.ID != i.ID {
		t.Fatal("el detalle autenticado debe conservar la inscripcion del jugador")
	}
	ruta := fmt.Sprintf("/inscripciones/%d/estado", i.ID)
	esperar(t, llamar(h, "PATCH", ruta, `{"estado":"pago_por_verificar","banco":"Pichincha","referencia":"DEMO-1"}`, jug), 200)
	esperar(t, llamar(h, "PATCH", ruta, `{"estado":"confirmada"}`, jug), 403)
	esperar(t, llamar(h, "PATCH", ruta, `{"estado":"confirmada"}`, org), 200)
	// Otra instancia del manejador ve los mismos datos guardados en PostgreSQL.
	h = servidorDB(db)
	var guardada Inscripcion
	if err := db.First(&guardada, i.ID).Error; err != nil || guardada.Estado != "confirmada" {
		t.Fatal("el estado no persistio")
	}
	var pago Pago
	if err := db.Where("inscripcion_id = ?", i.ID).First(&pago).Error; err != nil || !pago.Verificado {
		t.Fatal("pago sin verificar")
	}
	esperar(t, llamar(h, "PATCH", ruta, `{"estado":"cancelada"}`, org), 200)
	esperar(t, llamar(h, "PATCH", ruta, `{"estado":"confirmada"}`, org), 422)
	// Las dos cuentas conservan hashes; el perfil de negocio no contiene la clave.
	var cuenta sesion.Usuario
	if err := db.Where("nombre = ?", "ana").First(&cuenta).Error; err != nil {
		t.Fatal(err)
	}
	if cuenta.ClaveHash == "ana-2026" || !strings.HasPrefix(cuenta.ClaveHash, "$2") {
		t.Fatal("clave sin bcrypt")
	}
	esperar(t, llamar(h, "POST", "/login", `{"usuario":"ana","clave":"incorrecta"}`, ""), 401)
	esperar(t, llamar(h, "POST", "/login", `{"usuario":"nadie","clave":"incorrecta"}`, ""), 401)
	esperar(t, llamar(h, "POST", "/login", `{"usuario":`, ""), 400)
	esperar(t, llamar(h, "GET", "/eventos", "", ""), 200)
	esperar(t, llamar(h, "GET", "/eventos/25", "", ""), 200)
	if err := SembrarDemo(db); err != nil {
		t.Fatal(err)
	}
	if err := db.First(&guardada, i.ID).Error; err != nil || guardada.Estado != "cancelada" {
		t.Fatal("reiniciar la semilla no debe resetear reservas")
	}
}

func TestIntegracionUltimoCupoConcurrente(t *testing.T) {
	db := baseDePrueba(t)
	h := servidorDB(db)
	jug := login(t, h, "ana", "ana-2026")
	if err := db.Model(&Evento{}).Where("id = ?", 26).Update("cupos", 1).Error; err != nil {
		t.Fatal(err)
	}
	primera := crearReserva(t, h, 26, jug)
	u := Usuario{ID: 90, Nombre: "Otro jugador", Correo: "concurrencia@example.test", Rol: sesion.RolJugador}
	if err := db.Create(&u).Error; err != nil {
		t.Fatal(err)
	}
	otro, _ := tokensDePrueba.Emitir(sesion.Usuario{ID: 90, Rol: sesion.RolJugador})
	segunda := crearReserva(t, h, 26, otro)
	var wg sync.WaitGroup
	codigos := make(chan int, 2)
	for _, c := range []struct {
		i     Inscripcion
		token string
	}{{primera, jug}, {segunda, otro}} {
		wg.Add(1)
		go func(i Inscripcion, token string) {
			defer wg.Done()
			codigos <- llamar(h, "PATCH", fmt.Sprintf("/inscripciones/%d/estado", i.ID), `{"estado":"confirmada"}`, token).Code
		}(c.i, c.token)
	}
	wg.Wait()
	close(codigos)
	ok, conflicto := 0, 0
	for codigo := range codigos {
		if codigo == 200 {
			ok++
		} else if codigo == 409 {
			conflicto++
		} else {
			t.Fatalf("codigo concurrente inesperado: %d", codigo)
		}
	}
	if ok != 1 || conflicto != 1 {
		t.Fatalf("ultimo cupo: %d confirmadas y %d conflictos", ok, conflicto)
	}
	var n int64
	if err := db.Model(&Inscripcion{}).Where("evento_id = ? AND estado = ?", 26, "confirmada").Count(&n).Error; err != nil || n != 1 {
		t.Fatal("se sobrevendio el cupo")
	}
}

func TestIntegracionReglasDelNegocio(t *testing.T) {
	db := baseDePrueba(t)
	h := servidorDB(db)
	jug := login(t, h, "ana", "ana-2026")
	org := login(t, h, "luis", "luis-2026")
	// Modelo completo y tablas separadas de cuentas y usuarios.
	for _, tabla := range []string{"usuarios", "cuentas", "sedes", "eventos", "inscripciones", "pagos"} {
		if !db.Migrator().HasTable(tabla) {
			t.Fatal("falta tabla " + tabla)
		}
	}
	i := crearReserva(t, h, 26, jug)
	ruta := "/inscripciones/" + strconv.Itoa(int(i.ID)) + "/estado"
	esperar(t, llamar(h, "POST", "/eventos/26/inscripciones", "", jug), 409)
	esperar(t, llamar(h, "PATCH", ruta, `{"estado":"pago_por_verificar","banco":"Banco","referencia":"1"}`, jug), 409)
	esperar(t, llamar(h, "PATCH", ruta, `{"estado":"confirmada"}`, jug), 200)
	// Identidad ajena: aunque tenga el rol correcto no puede operar otra reserva.
	otro := Usuario{ID: 90, Nombre: "Otro", Correo: "otro@example.test", Rol: sesion.RolJugador}
	if err := db.Create(&otro).Error; err != nil {
		t.Fatal(err)
	}
	tokenAjeno, _ := tokensDePrueba.Emitir(sesion.Usuario{ID: 90, Rol: sesion.RolJugador})
	esperar(t, llamar(h, "PATCH", ruta, `{"estado":"cancelada"}`, tokenAjeno), 403)
	tokenOrgAjeno, _ := tokensDePrueba.Emitir(sesion.Usuario{ID: 90, Rol: sesion.RolOrganizador})
	esperar(t, llamar(h, "GET", "/eventos/26/inscripciones", "", tokenOrgAjeno), 403)
	// Una confirmada solo se cancela con al menos 24 horas.
	if err := db.Model(&Evento{}).Where("id = ?", 26).Update("inicio", time.Now().Add(time.Hour)).Error; err != nil {
		t.Fatal(err)
	}
	esperar(t, llamar(h, "PATCH", ruta, `{"estado":"cancelada"}`, org), 409)
	// No se acepta un estado prohibido y no se altera el anterior.
	esperar(t, llamar(h, "PATCH", ruta, `{"estado":"reservada"}`, org), 422)
	var actual Inscripcion
	db.First(&actual, i.ID)
	if actual.Estado != "confirmada" {
		t.Fatal("se modifico pese al rechazo")
	}
	// Evento pagado: no se puede saltar el comprobante.
	pagada := crearReserva(t, h, 25, jug)
	rp := fmt.Sprintf("/inscripciones/%d/estado", pagada.ID)
	esperar(t, llamar(h, "PATCH", rp, `{"estado":"confirmada"}`, jug), 409)
	esperar(t, llamar(h, "PATCH", rp, `{"estado":"pago_por_verificar"}`, jug), 422)
	esperar(t, llamar(h, "PATCH", rp, `{"estado":"pago_por_verificar","banco":"Banco","referencia":"2"}`, jug), 200)
	esperar(t, llamar(h, "PATCH", rp, `{"estado":"reservada"}`, org), 200)
	// Conflicto de cupo al confirmar, no se verifica el pago ni el estado.
	esperar(t, llamar(h, "PATCH", rp, `{"estado":"pago_por_verificar","banco":"Banco","referencia":"3"}`, jug), 200)
	db.Model(&Evento{}).Where("id = ?", 25).Update("cupos", 0)
	esperar(t, llamar(h, "PATCH", rp, `{"estado":"confirmada"}`, org), 409)
	var ultimo Pago
	db.Where("inscripcion_id = ?", pagada.ID).Order("id DESC").First(&ultimo)
	if ultimo.Verificado {
		t.Fatal("se verifico un pago sin cupo")
	}
	// FK reales: no se guarda una reserva cuyo evento no existe.
	if err := db.Create(&Inscripcion{EventoID: 99999, JugadorID: otro.ID, Estado: "reservada"}).Error; err == nil {
		t.Fatal("falta clave foranea")
	}
}
