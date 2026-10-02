# Hito 1 · Addendum técnico

**Pareja:** Sebastián Mendoza · Evelin Quijije
**Paralelo:** Aplicaciones Web II A

## A. Estructura del proyecto

Estructura objetivo del servidor de EventoSports; hoy el repositorio conserva la base de mesa de ayuda del curso.

```
.
├── main.go                        arranque: lee la configuración, conecta a PostgreSQL y levanta el servidor
├── go.mod / go.sum                módulo y dependencias (GORM, driver de PostgreSQL)
├── .env.example                   nombres de las variables de entorno, sin datos reales
├── .gitignore                     excluye .env y binarios
├── internal/
│   └── eventosports/
│       ├── modelos.go             structs Usuario, Sede, Evento, Inscripcion y Pago (sección 4 de la ficha)
│       ├── validaciones.go        listas cerradas (deporte, nivel, estado) y tabla de transiciones permitidas
│       ├── manejadores.go         los seis endpoints de la sección 7
│       ├── rutas.go               registro de rutas y verbos
│       └── manejadores_test.go    pruebas con httptest que corren sin base de datos
└── docs/
    ├── hito1_ficha_del_negocio.md
    ├── hito1_addendum.md
    ├── hito1_presentacion.pdf
    └── capturas y boceto (hito1_pruebas.png, hito1_boceto.png, hito1_respuesta_ok.png, hito1_respuesta_error.png)
```

## B. Configuración y secretos

| Variable | Para qué sirve | Ejemplo (sin datos reales) |
|----------|----------------|----------------------------|
| `DATABASE_URL` | Cadena de conexión a PostgreSQL (en Docker, base `mesa_ayuda` del curso o la de EventoSports) | `postgres://usuario:CAMBIE_ESTO@localhost:5432/eventosports` |
| `PORT` | Puerto en el que escucha el servidor | `8080` |
| `CLAVE_FIRMA` | Secreto para firmar sesiones cuando exista el inicio de sesión | `CAMBIE_ESTO` |

**Archivo de ejemplo:** `.env.example` en la raíz del repositorio. El `.env` real está en `.gitignore` y el repositorio no contiene credenciales.

## C. Pruebas

Las pruebas siguen las reglas del kit de la semana 4: todo lo que se valida responde antes de llegar a la línea que usa la base de datos.

| Prueba | Qué caso cubre |
|--------|----------------|
| 1. `TestCrearEvento_JSONRoto` | `POST /eventos` con cuerpo mal formado responde 400. |
| 2. `TestCambiarEstado_FueraDeLaLista` | `PATCH /inscripciones/118/estado` con `{"estado":"vip"}` responde 422. |
| 3. `TestTransicion_CanceladaAConfirmada` | La tabla de transiciones de `validaciones.go` rechaza `cancelada → confirmada` (el manejador responde 409). |
| 4. `TestCrearEvento_CuposCero` | `POST /eventos` con `cupos: 0` responde 422. |
| 5. `TestVerEvento_IdNoNumerico` | `GET /eventos/abc` responde 400. |
| 6. `TestCrearEvento_RolJugador` | `POST /eventos` con `X-Rol: jugador` responde 403. |

Esta tabla es el plan de pruebas de EventoSports (la Unidad 1 pidió cinco como mínimo).

**Captura de `go test ./...`:** ![Pruebas](hito1_pruebas.png)

## D. Boceto de la pantalla principal

Pantalla **Panel del organizador**, que consume el listado `GET /eventos/25/inscripciones`. Cada fila muestra el jugador, el **estado** de su inscripción (insignia de color: reservada, pago por verificar, confirmada o cancelada), el banco y la referencia de la transferencia y el monto. Los botones **Verificar**, **Rechazar** y **Cancelar** llaman a `PATCH /inscripciones/{id}/estado`. La pantalla de los jugadores, *Explorar eventos*, lista eventos con `GET /eventos` y no muestra estado porque los eventos no lo tienen: el estado de la inscripción se ve en este panel y en el *Detalle del evento*.

![Boceto](hito1_boceto.png)

## E. Diagrama de secuencia del caso de uso principal

Caso: un jugador reserva cupo en un evento de pago y el organizador verifica la transferencia.

```mermaid
sequenceDiagram
    actor Jugador
    actor Organizador
    Jugador->>Detalle: toca "Apuntarme"
    Detalle->>API: POST /eventos/25/inscripciones
    API->>BD: valida evento, cupos y que no esté inscrito, y guarda la inscripción
    BD-->>API: ok
    API-->>Detalle: 201 con inscripción en estado reservada
    Jugador->>Detalle: transfiere y escribe banco y referencia
    Detalle->>API: PATCH /inscripciones/118/estado {"estado":"pago_por_verificar","banco":"Banco Pichincha","referencia":"4839201"}
    API->>BD: valida la transición y guarda el pago con monto, banco y referencia
    BD-->>API: ok
    API-->>Detalle: 200 con estado pago_por_verificar
    Organizador->>Panel: revisa su banco y toca "Verificar"
    Panel->>API: PATCH /inscripciones/118/estado {"estado":"confirmada"}
    API->>BD: valida la transición, el dueño del evento y los cupos, y marca el pago como verificado
    BD-->>API: ok
    API-->>Panel: 200 con estado confirmada
```

## F. Capturas de respuestas

Las capturas se toman con Bruno o `curl` sobre el servidor corriendo, con el código HTTP y el cuerpo JSON a la vista. Ejemplos de petición:

```
curl -i -X POST http://localhost:8080/eventos -H "X-Usuario-ID: 7" -H "X-Rol: organizador" -H "Content-Type: application/json" -d '{"titulo":"Fútbol 7 mixto","deporte":"futbol","nivel":"casual","inicio":"2026-10-15T20:00:00-05:00","cupos":14,"precio":2.5,"sede_id":3}'
curl -i -X POST http://localhost:8080/eventos -H "X-Usuario-ID: 7" -H "X-Rol: organizador" -H "Content-Type: application/json" -d '{"titulo":"Fútbol 7 mixto","deporte":"futbol","nivel":"casual","inicio":"2026-10-15T20:00:00-05:00","cupos":0,"precio":2.5,"sede_id":3}'
```

**Caso correcto:** ![Respuesta correcta](hito1_respuesta_ok.png)

**Caso con error de validación:** ![Respuesta con error](hito1_respuesta_error.png)
