# EventoSports · Hito 2 y taller 11

Sebastián Mendoza y Evelin Quijije · Grupo E · Aplicaciones Web II A.

## Taller 11 · la API en un contenedor

Rama `taller11`. Parte del código del Hito 2. Las evidencias están en `docs/taller11/`, y las decisiones D1 a D9, en `docs/decisiones.md`.

| Parte | Qué hay | Evidencia |
|---|---|---|
| Fase 0 | Línea de base con `go run` y respuesta completa de la petición 3 | `docs/taller11/peticiones.md` |
| 1. La imagen | `servidor/Dockerfile` de dos etapas y `servidor/.dockerignore`. La imagen pesa 39.6 MB, la de compilación 1.63 GB, no entran `.env` ni `.git`, y arranca como el usuario 65532 | `imagenes.txt`, `capas.txt`, `contexto.txt`, `usuario.txt` |
| 2. El contenedor | Arranca sin `.env`, con `--env-file .env.docker` y `-p 8080:8080`, y responde los mismos seis códigos. `main.go` registra cada petición en el log | `1_sin_variables.txt`, `2_local.txt`, `logs_contenedor.txt`, `peticiones.md` |
| Reto | `servidor/compose.yaml` levanta la base y la API con un solo comando, sin `host.docker.internal` ni claves escritas | `compose.txt`, `peticiones.md` |

Comprobado el 12 de noviembre de 2026: las seis peticiones dan 401, 200, 200, 403, 200 y 422 con `go run`, en el contenedor y con Compose. Compose arrancó tres veces seguidas desde una base vacía, sin reinicios de la API, y los datos sobrevivieron a un `docker compose down`.

Desde `servidor/`:

```powershell
docker build -t eventosports .
docker run -d --name eventosports-api -p 8080:8080 --env-file .env.docker eventosports
docker compose up -d --build
```

`.env.docker` es una copia de `.env` con `host.docker.internal` en lugar de `localhost`, y Git lo ignora. Compose toma las claves de `servidor/.env`. **`docker compose down -v` borra la base de Compose**: úsese solo para empezar desde cero.

La única línea de Go que se agregó es `r.Use(middleware.Logger)` en `main.go`: sin ella, el log del contenedor no mostraba las peticiones, y el taller pide esa prueba.

La entrega técnica reúne PostgreSQL, autenticación JWT, permisos, estados y pruebas del taller S7 con la separación en capas de la semana 8. La demostración usa una colección de Bruno preparada para responder **401 → 200 → 200 → 403 → 200 → 422**.

El enunciado oficial pide una demostración en vivo de cinco minutos y especifica que **no hay diapositivas**. Si un integrante falta, la persona presente realiza las cinco demostraciones y responde las dos preguntas.

## Inicio rápido

Desde la raíz, en una terminal:

```powershell
cd servidor
powershell -ExecutionPolicy Bypass -File scripts/preparar-entorno.ps1
docker compose up -d --build
```

Eso levanta la base y la API juntas, en `http://localhost:8080`.

Para trabajar con `go run .`, en cambio, hace falta PostgreSQL 16 instalado en el equipo, en `localhost:5432`, con el usuario y la base `eventosports`. La primera vez se crean así, desde `servidor/`, con la clave del usuario `postgres`:

```powershell
$clave = ((Select-String -LiteralPath .env -Pattern '^POSTGRES_PASSWORD=').Line -split '=',2)[1]
& 'C:\Program Files\PostgreSQL\16\bin\psql.exe' -U postgres -h localhost -d postgres -c "CREATE ROLE eventosports LOGIN PASSWORD '$clave'" -c "CREATE DATABASE eventosports OWNER eventosports"
```

`.env` guarda la conexión y el secreto JWT local y está excluido de Git.

Con `DEMO_SEED=true`: Ana (`ana` / `ana-2026`) es jugador y Luis (`luis` / `luis-2026`) es organizador. Son credenciales inventadas de la demostración; las contraseñas se guardan con bcrypt.

## Verificación completa

Con la API encendida, desde otra terminal en la raíz:

```powershell
powershell -ExecutionPolicy Bypass -File servidor/scripts/verificar-hito2.ps1
```

El script valida el contrato, corre las pruebas, revisa y compila Go, y ejecuta la colección real de Bruno. Exige configuración de base para que la integración no se omita silenciosamente.

Si el puerto 8080 ya está ocupado, iniciar la API con un puerto libre, por ejemplo `$env:PORT='18080'; go run .` desde `servidor/`, y pasar su dirección al verificador: `powershell -ExecutionPolicy Bypass -File servidor/scripts/verificar-hito2.ps1 -ApiBaseUrl http://localhost:18080`.

También se puede abrir `docs/hito2_bruno/` en el cliente de Bruno, seleccionar **Local**, enviar la petición **0** antes del turno y luego enviar **1 a 6** en orden. La preparación crea registros nuevos y conserva los IDs y tokens solo en variables de ejecución. No hay que pegar tokens ni modificar archivos durante el turno. Si se repite la 5 sobre el mismo registro ya confirmado devuelve 422; ejecutar de nuevo la 0 prepara otro ensayo.

## Entregables

| Archivo o carpeta | Contenido |
|---|---|
| `docs/openapi.yaml` | Contrato completo, seguridad y operaciones futuras marcadas `(plan)` |
| `docs/api.html` | Documentación generada del contrato |
| `redocly.yaml` | Reglas del curso, con `security-defined` activo |
| `docs/hito2_hoja.md` y `docs/hito2_hoja.html` | Hoja breve y versión imprimible en una página A4 |
| `docs/hito2_bruno/` | Siete peticiones 0–6, entorno Local y comprobaciones |
| `docs/hito2_capturas/` | Imágenes del informe de respuestas reales y pruebas; reporte JSON sin JWT |
| `docs/decisiones.md` | Sección «Semana 8 · estructura», alcance y decisiones |
| `docs/hito2_guion.md` | Preparación, cinco minutos y seis preguntas para practicar |
| `docs/README_taller_s7.md` | Explicación histórica del taller S7, antes de las capas |

## Capas

- `servidor/internal/eventosports/`: HTTP y rutas; no consulta la base.
- `servidor/internal/negocio/`: casos de uso, interfaces y reglas; no conoce HTTP ni GORM.
- `servidor/internal/datos/`: consultas, transacciones y migración PostgreSQL.
- `servidor/internal/modelo/`: cinco entidades y relaciones.
- `servidor/internal/sesion/`: código de sesión del docente, bcrypt, JWT y guardias.
- `servidor/internal/configuracion/`: lectura de variables.

Para la demostración 5, abrir **`eventosports/estado.go`** y **`negocio/estados.go`**. El primero recibe y responde; el segundo decide las transiciones sin HTTP. La conexión entre interfaces e implementación se arma en `main.go`.

Se añadió `GET /inscripciones/{id}` para consultar individualmente la entidad principal. Las rutas futuras de perfiles, sedes y pagos solo se documentaron: no se implementó una tercera entidad. El PATCH conserva los permisos por transición de la ficha; verificar un pago es exclusivo del organizador dueño.

## Qué se comprobó

- `go test ./... -count=1`: pruebas de seguridad, tres integraciones PostgreSQL y regla pura aprobadas, sin omitir la integración local.
- `go vet ./...` y `go build ./...`: correctos.
- Redocly: contrato válido sin errores ni avisos.
- Bruno CLI 4.3.0: **7 peticiones aprobadas y 9 comprobaciones correctas** contra el servidor reorganizado.
- Los respaldos visuales se generaron a partir de esa ejecución real; su README explica el origen y la ocultación de JWT. No son capturas de la interfaz del cliente Bruno.

## Antes de entregar

La parte técnica está preparada para ensayar. Todavía deben revisar y explicar el código, ensayar los cinco minutos con la computadora de la demostración, crear los commits propios de ambos integrantes y comprobar el push al repositorio de la organización antes del cierre. No se hicieron commits ni push automáticamente.

Para Web II A, según el enunciado: cierre **29 de octubre de 2026 a las 23:59** y demostración **30 de octubre a las 09:00**, según el orden del docente. Una ejecución rápida del CLI no reemplaza el ensayo oral de ambos integrantes.

## Uso de IA

La integración S7, las capas, las pruebas y los materiales de preparación se realizaron con asistencia de Codex. El código de tokens y guardia procede del material docente. La declaración y el detalle de decisiones están en `docs/decisiones.md`.
