> Documento historico del taller S7, anterior a la separacion en capas. Para el estado actual consultar README.md en la raiz.

# EventoSports · Taller integrador de la semana 7

Proyecto de Sebastián Mendoza y Evelin Quijije · Aplicaciones Web II A.

## Resultado

La API pasó de guardar datos en memoria y confiar en `X-Usuario-ID` / `X-Rol` a guardar en PostgreSQL y obtener la identidad de un JWT firmado. Se completó la integración de los materiales S7 disponibles: estados, sesión, permisos, pruebas y contrato OpenAPI.

**Verificado el 26 de octubre de 2026:** compilación, `go vet ./...`, pruebas de Go con integración real en PostgreSQL, contrato sin errores ni avisos y las seis peticiones contra el servidor encendido.

Este README explica el taller realizado. La organización final en capas y los archivos específicos de la demostración del Hito 2 siguen siendo un trabajo posterior; se detallan al final. El enunciado completo del taller S7 no estaba entre los archivos recibidos: el alcance se contrastó con el paquete de código, sus instrucciones y los requisitos S7 enumerados en el Hito 2.

## 1. Qué había y qué cambió

| Antes | Después |
|---|---|
| Mapas en memoria; se perdían al cerrar | Tablas PostgreSQL y volumen persistente |
| Identidad elegida por el cliente en encabezados | Identidad y rol tomados del token verificado |
| Sin inicio de sesión | `POST /login` con bcrypt y JWT |
| Enrutamiento manual | Chi y grupo de rutas con autenticación |
| Estados prohibidos con 409 | Transición prohibida con **422**, como el taller |
| Fechas de ejemplo ya vencidas | Al crear los ejemplos, eventos a 30 y 31 días del momento de creación |
| Pruebas solo en memoria | Pruebas sin base para la puerta de seguridad y pruebas reales de PostgreSQL |

Se conservaron las cinco entidades del negocio y sus seis rutas. No se añadieron registro ni CRUD de cuentas. Las fechas de ejemplos ya existentes no se cambian en cada reinicio: para nuevos ensayos, el script crea eventos nuevos con fecha futura.

## 2. Archivos del docente utilizados

| Material entregado | Integración realizada |
|---|---|
| `tokens.go` | Copiado a `servidor/internal/sesion/`, sin cambios de lógica |
| `guardia.go` | Copiado a la misma carpeta, sin cambios de lógica |
| `usuarios.go` | Roles `jugador` y `organizador`; tabla `cuentas`, como indica el material cuando ya existe `usuarios` |
| `sesion_ayuda_test.go` | Adaptado al paquete `eventosports`, al módulo y al rol organizador |
| `sin_token_test.go` | Lista de las cuatro operaciones protegidas del proyecto |
| `estado.go` | Adaptado a Inscripción y a las reglas existentes de EventoSports |
| Esqueleto y fragmentos OpenAPI | Integrados en `docs/openapi.yaml` |
| Configuración Redocly | `redocly.yaml` en la raíz, con `security-defined: error` |

**Por qué `estado.go` necesita más adaptación:** el ejemplo del docente cambia un estado de un ticket. Nuestro proyecto ya tenía permisos por transición, dueño de la inscripción, pagos, cupos y plazos. Copiar literalmente el ejemplo eliminaría esas reglas. Por eso se conserva su flujo de validar → consultar → comprobar transición → guardar, añadiendo las condiciones del negocio y una transacción.

## 3. Base de datos y Docker

Se crea un solo servicio para este proyecto: **PostgreSQL 16**.

| Elemento | Valor |
|---|---|
| Contenedor | `eventosports-s7-db-1` |
| Base | `eventosports` |
| Usuario de conexión | `eventosports` |
| Dirección desde Windows | `localhost:5433` |
| Puerto dentro del contenedor | `5432` |
| Volumen | `eventosports-s7_postgres_data` |
| Configuración | `servidor/compose.yaml` |

La API de Go corre en Windows, fuera de Docker. Los servicios de Gastos App/Supabase y facturación que ya existían no pertenecen a este proyecto. Al iniciar Docker Desktop, algunos servicios anteriores se reactivaron automáticamente; no se cambiaron sus bases ni su configuración.

Al iniciar la API se crean estas tablas y sus relaciones:

- `usuarios`: perfiles del negocio.
- `sedes`: lugares de los eventos.
- `eventos`: relacionados con sede y organizador.
- `inscripciones`: relacionadas con evento y jugador.
- `pagos`: relacionados con una inscripción.
- `cuentas`: nombres, roles y hashes para el login; separada de `usuarios`.

También se creó un índice que impide dos inscripciones activas del mismo jugador en un evento. Una inscripción cancelada permite crear otra. Las operaciones que afectan cupos bloquean el evento dentro de una transacción para evitar confirmar dos reservas sobre el último cupo.

## 4. Cómo iniciar el proyecto

Requisitos: Go, Docker Desktop funcionando y Node.js/npm para validar el contrato y ejecutar el recorrido HTTP. Se verificó con Go 1.27.1 y PostgreSQL 16. Las versiones de las bibliotecas están fijadas en `servidor/go.mod` y `go.sum`.

Desde la raíz del repositorio, en la terminal de Visual Studio Code:

```powershell
cd servidor
powershell -ExecutionPolicy Bypass -File scripts/preparar-entorno.ps1
docker compose up -d --wait db
go run .
```

El script de preparación crea `.env` con claves aleatorias si no existe. Si ya existe, lo conserva. El servidor queda en `http://localhost:8080`. Para las verificaciones realizadas ya se generó la configuración local y se inició PostgreSQL.

En otra terminal se ejecutan las pruebas o peticiones; la primera terminal debe seguir abierta para mantener el servidor funcionando.

### Variables de configuración

| Variable | Función |
|---|---|
| `PORT` | Puerto HTTP; por defecto 8080 |
| `DATABASE_URL` | Conexión PostgreSQL |
| `POSTGRES_PASSWORD` | Clave local del contenedor |
| `JWT_SECRET` | Firma del token; el arranque exige al menos 32 caracteres |
| `DEMO_SEED` | Con `true`, prepara los dos usuarios y los datos de ejemplo |

`servidor/.env.example` muestra los nombres y valores de reemplazo. `.env` está excluido de Git; las claves locales de la base y la firma no se incluyeron en este README.

Para usar PostgreSQL instalado directamente en Windows, se puede cambiar `DATABASE_URL` y omitir el comando de Docker. La base debe existir y el usuario debe poder crear las tablas.

Para detener solo la base de EventoSports:

```powershell
docker compose stop db
```

Los datos permanecen en el volumen. No es necesario eliminar volúmenes para ensayar.

## 5. Login y JWT

Con `DEMO_SEED=true` se preparan estas cuentas inventadas:

| Usuario | Clave de ejemplo | Rol |
|---|---|---|
| `ana` | `ana-2026` | `jugador` |
| `luis` | `luis-2026` | `organizador` |

Las contraseñas se guardan como hashes bcrypt, nunca directamente. La semilla sigue el archivo del docente: al reiniciar actualiza las cuentas de ejemplo y sus hashes; no reinicia las inscripciones ni los pagos.

Las cuentas de ejemplo y sus perfiles de negocio se crean con el mismo ID para que el `sub` del JWT identifique al dueño. No se supone que Ana tenga el ID 12 ni Luis el 7 del antiguo servidor en memoria. Los manejadores siempre usan la identidad que verifica el middleware.

Ejemplo en PowerShell:

```powershell
$sesion = Invoke-RestMethod -Method Post -Uri http://localhost:8080/login `
  -ContentType application/json `
  -Body '{"usuario":"luis","clave":"luis-2026"}'

$cabeceras = @{ Authorization = "Bearer $($sesion.token)" }
Invoke-RestMethod -Uri http://localhost:8080/eventos/25/inscripciones -Headers $cabeceras
```

El token dura **una hora** y lleva `sub` (ID), `nombre`, `rol`, `iat` (emisión) y `exp` (vencimiento). Se firma con HS256. Sus datos están codificados y pueden leerse: la firma detecta modificaciones. Un token cambiado, vencido, sin vencimiento o firmado con otro secreto responde **401** en una ruta protegida.

## 6. Matriz de rutas y permisos

| Operación | Sin token | Jugador | Organizador |
|---|---|---|---|
| `POST /login` | Sí, con credenciales | Sí | Sí |
| `GET /eventos` | Sí | Sí | Sí |
| `GET /eventos/{id}` | Sí | Sí; con token incluye su última inscripción | Sí |
| `POST /eventos` | 401 | 403 | Sí |
| `POST /eventos/{id}/inscripciones` | 401 | Sí, para sí mismo | 403 |
| `GET /eventos/{id}/inscripciones` | 401 | 403 | Solo sus eventos |
| `PATCH /inscripciones/{id}/estado` | 401 | Solo sus inscripciones y transiciones | Solo sus eventos y transiciones |

La protección se registra en `servidor/internal/eventosports/rutas.go`, en `r.Use(tokens.Autenticar)`. `SoloRol` restringe las operaciones de un solo rol. El PATCH acepta ambos roles y después comprueba propiedad y transición. Cambiar `X-Rol` o `X-Usuario-ID` ya no concede permisos.

El detalle de un evento sigue siendo público. Si se envía `Authorization`, se valida antes de mostrar información personal; si ese token no sirve, responde 401.

## 7. Estados de Inscripción

| De | A | Quién | Condición |
|---|---|---|---|
| `reservada` | `pago_por_verificar` | Jugador dueño | Evento pagado y banco/referencia presentes |
| `reservada` | `confirmada` | Jugador dueño | Evento gratuito y cupo disponible |
| `pago_por_verificar` | `confirmada` | Organizador dueño del evento | Cupo disponible; marca el último pago verificado |
| `pago_por_verificar` | `reservada` | Organizador dueño | Rechazo del comprobante |
| `reservada` o `pago_por_verificar` | `cancelada` | Jugador u organizador dueño | El evento no ha empezado |
| `confirmada` | `cancelada` | Jugador u organizador dueño | Faltan al menos 24 horas |
| `cancelada` | Cualquier estado | Nadie | Estado final: responde 422 |

La tabla de transiciones está en `validaciones.go`; el flujo HTTP y transaccional, en `estado.go`. El monto de un pago sale del precio del evento, no de un valor que envíe el cliente.

### Códigos que hay que distinguir

- **400:** JSON mal formado o ID inválido.
- **401:** falta autenticación válida o las credenciales no coinciden.
- **403:** se conoce al usuario, pero su rol o propiedad no permite la acción.
- **404:** el recurso no existe.
- **409:** inscripción activa repetida, falta de cupos, pago previo requerido o conflicto de plazos.
- **422:** estado desconocido, transición prohibida o datos que incumplen validaciones.
- **500:** error interno de almacenamiento o sesión.

La transición prohibida antes usaba 409 y ahora usa 422. Se actualizó el contrato y las pruebas; la documentación del hito 1 se conserva como registro histórico.

## 8. Cómo comprobarlo

Desde `servidor/`:

```powershell
go test ./... -count=1 -v
go vet ./...
go build ./...
node scripts/comprobar-s7.mjs
```

El último comando necesita la API encendida. Obtiene tokens nuevos, crea dos eventos y dos inscripciones de prueba, prepara una pendiente de verificar y otra cancelada, y envía:

| Paso | Petición | Esperado y comprobado |
|---|---|---|
| 1 | Listado de inscripciones sin token | 401 |
| 2 | Login de Luis | 200 y token |
| 3 | Mismo listado con token de Luis | 200 |
| 4 | Ana intenta verificar el pago | 403 |
| 5 | Luis verifica el mismo pago | 200 |
| 6 | Luis intenta pasar de cancelada a confirmada | 422 |

El script es repetible porque crea sus propios registros; no borra los datos existentes. En las peticiones de preparación ya inicia sesión como ambos roles para crear esos datos; el login del paso 2 vuelve a hacerse durante el recorrido. Nunca imprime los tokens. Admite una URL alternativa como primer argumento y una ruta de reporte JSON como segundo.

### Qué cubren las pruebas

- Cuatro operaciones protegidas devuelven 401 antes de consultar la base si no hay token.
- Un cliente no puede inventar su identidad con los encabezados antiguos.
- Se rechazan tokens manipulados, vencidos, sin `exp`, con otra firma o mal formados.
- Se comprueban JSON, IDs, filtros y cupos inválidos antes de acceder a los datos.
- Login correcto e incorrecto; hashes bcrypt y tabla de cuentas separada.
- Flujo de reserva, pago, verificación, cancelación y transición prohibida.
- Propiedad de inscripciones y eventos; rechazo de operaciones ajenas.
- Persistencia del estado y del pago, incluido después de volver a ejecutar la semilla.
- Reglas de anticipación, pago previo, reservas repetidas y claves foráneas reales.
- Dos confirmaciones concurrentes sobre el último cupo: una responde 200 y la otra 409.

Las pruebas de PostgreSQL crean un esquema temporal propio y lo eliminan al terminar. No borran ni truncan las tablas de la aplicación. Usan `TEST_DATABASE_URL` si está definido; de lo contrario leen `DATABASE_URL` de `servidor/.env`. Si no hay ninguna URL, se omiten explícitamente las pruebas de base, aunque las de seguridad sin base siguen corriendo. **En esta entrega se ejecutaron las tres pruebas de integración con PostgreSQL y ninguna fue omitida.**

## 9. Contrato y documentación

Desde la raíz del repositorio:

```powershell
npx @redocly/cli lint docs/openapi.yaml --format=stylish
npx @redocly/cli build-docs docs/openapi.yaml --output docs/api.html
```

`docs/openapi.yaml` describe las seis rutas originales y el login, esquemas, cuerpos, seguridad JWT y respuestas. `security` protege por defecto; las consultas públicas y el login declaran sus excepciones. El login ya no se marca como plan porque está implementado. `docs/api.html` se generó a partir del contrato final.

## 10. Dónde está cada cosa

```text
README.md                          Esta explicación
redocly.yaml                       Reglas del contrato
docs/openapi.yaml                  Contrato S7
docs/api.html                      Documentación generada
servidor/
  main.go                          Configuración, conexión, migración y arranque
  compose.yaml                     Un PostgreSQL para EventoSports
  .env.example                     Ejemplo de configuración sin secretos reales
  go.mod / go.sum                   Dependencias fijadas
  scripts/
    preparar-entorno.ps1            Genera .env local con claves aleatorias
    comprobar-s7.mjs                Preparación y seis peticiones reales
  internal/
    configuracion/configuracion.go Lectura de variables y .env
    sesion/
      usuarios.go                  Cuentas, bcrypt y login
      tokens.go                    Emisión y verificación de JWT
      guardia.go                   Autenticación y permisos por rol
    eventosports/
      modelos.go                   Cinco modelos y relaciones
      base.go                      Migración y datos de demostración
      rutas.go                     Rutas públicas y protegidas
      validaciones.go              Listas y tabla de transiciones
      manejadores.go               Operaciones HTTP y consultas
      estado.go                    Cambio de estado transaccional
      sesion_ayuda_test.go          Ayudas del taller para pruebas
      sin_token_test.go            Prueba de las rutas protegidas
      manejadores_test.go          Regresión, seguridad e integración
```

## 11. Qué falta para el Hito 2

El taller S7 aporta la base funcional. Antes de afirmar que el Hito 2 está terminado todavía hay que:

1. Aplicar la guía de semana 8: separar consultas a datos de los manejadores y aislar por completo la regla de negocio de HTTP.
2. Explicarlo en `docs/decisiones.md`, sección «Semana 8 · estructura».
3. Preparar `docs/hito2_hoja.md`, la colección `docs/hito2_bruno/` y las capturas `docs/hito2_capturas/`.
4. Revisar el contrato contra cualquier cambio del mapa del hito 1 y marcar como «(plan)» las operaciones futuras que corresponda agregar.
5. Adaptar y ensayar las seis peticiones de la demostración en Bruno. Nuestro PATCH usa permisos por transición, no un único rol para todos los cambios; el caso elegido es verificar un pago, exclusivo del organizador.
6. Ensayar con reloj, preparar las respuestas individuales y tener commits reales de ambos integrantes.
7. Subir el repositorio de la organización antes del cierre. No se hizo commit ni push automáticamente.

## 12. Uso de IA

La adaptación de PostgreSQL, sesión JWT, rutas, pruebas, contrato, scripts y este README se realizó con asistencia de Codex. Los archivos `tokens.go` y `guardia.go` provienen del material del docente; `usuarios.go` y las ayudas de pruebas se adaptaron como se explica arriba. La pareja debe revisar y poder explicar las decisiones y el código antes de demostrar.
