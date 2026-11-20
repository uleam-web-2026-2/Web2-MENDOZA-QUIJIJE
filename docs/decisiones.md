# Decisiones de EventoSports

## Semana 8 · estructura

Elegimos separar HTTP, negocio, datos y modelo en paquetes dentro de `servidor/internal`. Inscripción sigue siendo la entidad principal porque su estado representa la reserva y la verificación de pago; Evento es la entidad relacionada que ya tenía endpoints en el hito 1.

| Carpeta | Responsabilidad | Qué no hace |
|---|---|---|
| `eventosports/` | Rutas, lectura de JSON e ID, sesión y traducción de errores a códigos HTTP | No ejecuta consultas ni decide las transiciones |
| `negocio/` | Casos de uso, validaciones, propiedad, permisos por transición e interfaces del almacén | No importa `net/http`, GORM ni el adaptador de datos |
| `datos/` | GORM, PostgreSQL, consultas, bloqueos, transacciones, migración y semilla | No interpreta peticiones HTTP |
| `modelo/` | Los cinco structs del negocio y sus relaciones | No contiene manejadores ni casos de uso |
| `sesion/` | Los archivos de autenticación del docente: bcrypt, JWT y guardias | No decide las transiciones de Inscripción |
| `configuracion/` | Variables de entorno y lectura de `.env` | No contiene credenciales reales en el código |

El paquete `sesion` conserva la estructura del código entregado por el docente; la separación exigida para la entidad principal está en los cuatro primeros paquetes. No pretendemos que el manejador de login ya sea una arquitectura de capas independiente.

```mermaid
flowchart LR
    H[HTTP: eventosports] --> N[Casos de uso: negocio]
    N --> I[Interfaces Almacen y Operacion]
    D[Adaptador datos: GORM] -. implementa .-> I
    D --> P[(PostgreSQL)]
```

`main.go` conecta las piezas: entrega un `datos.AlmacenGORM` al servicio de negocio y entrega ese servicio al manejador HTTP. Las interfaces se declaran en `negocio/contratos.go`, porque el consumidor define qué necesita de la persistencia. El negocio no importa `datos`.

### Dos archivos que mostraremos

1. **`servidor/internal/eventosports/estado.go`**: recibe el ID y el JSON, llama a `Servicio.CambiarEstado` y escribe la respuesta. No contiene `DB`, `Where`, `First`, `Save` ni una transacción GORM.
2. **`servidor/internal/negocio/estados.go`**: `ValidarCambio` recibe inscripción, evento, actor, estado solicitado, ocupación y hora. Devuelve una decisión de negocio. No recibe `http.Request`, no escribe `http.ResponseWriter` y no usa códigos HTTP ni SQL.

La coordinación de la transacción está en `negocio/servicio.go`; su ejecución concreta y los bloqueos están en `datos/almacen_gorm.go`. El servicio bloquea primero el evento y después la inscripción. La prueba de concurrencia comprueba que dos jugadores no puedan confirmar el último cupo a la vez.

### Errores

El negocio devuelve tipos: `Regla`, `Prohibido`, `Conflicto` o `ErrNoExiste`. La capa HTTP los convierte en 422, 403, 409 y 404 respectivamente. El JSON roto y el ID inválido producen 400 en HTTP. Así, cambiar un código de respuesta no exige modificar la regla de estados.

### Qué cambiaría con otra base

Cambiaríamos el adaptador de `datos/`, las migraciones y la conexión en `main.go`; escribiríamos una implementación de las mismas interfaces. Las reglas de `negocio/estados.go`, los casos de uso y las rutas HTTP podrían mantenerse. La prueba pura de estados sigue ejecutándose sin base de datos.

## Alcance y contrato del Hito 2

- Se conservaron las seis operaciones que ya existían en el mapa del hito 1.
- Se añadió `GET /inscripciones/{id}` para cumplir el mínimo de consultar individualmente la entidad principal. Solo permite al jugador dueño o al organizador de su evento.
- Los modelos de Usuario, Sede, Evento, Inscripción y Pago y todas sus relaciones se migran. `cuentas` es una tabla adicional de sesión, separada de `usuarios`.
- Se documentaron como **`(plan)`** el listado y alta de perfiles, sedes e historial de pagos. No se programaron esos endpoints. Es una extensión del mapa futuro, no una afirmación de que esas rutas ya funcionan.
- No se programaron edición, borrado ni registro de cuentas. Cancelar una inscripción es un cambio de estado.

## Permisos por operación del negocio

El PATCH de Inscripción acepta ambos roles porque la ficha ya permite que un jugador cancele o envíe un comprobante y que un organizador verifique pagos. **La operación demostrada es pasar de `pago_por_verificar` a `confirmada`, exclusiva del organizador dueño.** Ana recibe 403; Luis recibe 200 con el mismo cuerpo. No restringimos todo el PATCH a organizador porque eso quitaría acciones legítimas del jugador.

La transición `cancelada → confirmada` está prohibida aun para Luis: responde 422. La petición está bien formada y el actor es dueño del evento; falla la regla, no la sintaxis ni la autenticación.

## Datos reproducibles y colección

La petición 0 de Bruno hace la preparación antes del turno: inicia ambas sesiones, crea dos eventos futuros, crea y cancela una reserva de respaldo y crea otra cuyo comprobante queda pendiente. La respuesta HTTP principal de la petición 0 es la creación de esa segunda inscripción, 201; su script posterior registra el comprobante. Cada petición auxiliar comprueba su resultado.

Los IDs y los tokens se conservan únicamente en variables de ejecución `bru.setVar`. No se escriben JWT en los archivos `.bru` ni en el entorno versionado. La petición 2 vuelve a iniciar sesión en vivo y obtiene el token usado en 3, 5 y 6. Las peticiones 4 y 5 comparten ID y cuerpo. La 6 usa otro registro ya cancelado. Ejecutar de nuevo la 0 prepara datos nuevos sin borrar datos anteriores.

## Uso de IA

La separación en capas, las pruebas, el contrato, los scripts, la colección y esta documentación se prepararon con asistencia de Codex. Los archivos de tokens y guardia provienen del material docente. La pareja debe revisar las decisiones y explicar el código con sus propias razones. No se fabricaron commits de ningún integrante ni se hizo push automáticamente.

## Semana 12 · la seguridad de la API

### D1 · Lo que encontramos y cerramos
La búsqueda con `docs/taller12/patrones.txt` dio 27 líneas: 21 arman SQL y 6 no (un grupo de rutas, dos mensajes de error, el correo de la siembra, la lectura de `r.URL.Query()` y la dirección del servidor). El texto del cliente llega al SQL por cuatro lugares: `?deporte=`, `?nivel=`, `?ciudad=` y el `usuario` del login. Los cuatro van como parámetro (`?`), y `deporte` y `nivel` además se comparan con una lista y responden 422 si no están. Los `{id}` se convierten con `strconv.ParseUint` antes de la consulta, y si no son números responden 400. El único `+` dentro de un SQL (`datos/base.go:55`) pega nombres de tabla fijos del código en la siembra, no texto del cliente. No había nada vulnerable, así que no se cambió código. La fila por fila está en `docs/taller12/auditoria.md`.

La prueba del paso 1.4 (`inyeccion.txt`): `?ciudad=Manta` devolvió 4 eventos; `?ciudad=' OR 1=1 --` devolvió `[]`, porque PostgreSQL buscó una ciudad que se llama literalmente así; `?deporte=' OR 1=1 --` dio 422; y `/inscripciones/1%20OR%201=1` dio 400. El mismo ataque quedó como prueba automática en `internal/eventosports/inyeccion_test.go`.

### D2 · El origen que puede leer nuestras respuestas
Nombramos `http://localhost:5173`, donde va a correr el front con Vite (todavía no lo tenemos). No usamos `"*"` porque cualquier página que abra un usuario podría llamar a la API desde su navegador y leer las respuestas. Con el origen nombrado, solo nuestro front recibe ese permiso. `AllowedMethods` lleva solo GET, POST y PATCH, los métodos que usa la API.

CORS va justo debajo de `chi.NewRouter()`, antes que todo, porque la petición previa (`OPTIONS`) no trae token: si pasara antes por el guardia recibiría 401, y metida en un `r.With` de las rutas daría 405 porque `OPTIONS` no tiene ruta propia. Por eso tampoco aparece en el registro de `middleware.Logger`, que va debajo: CORS la responde antes.

Desde `http://malo.example`, la previa respondió 200 sin ninguna cabecera `Access-Control-…`, y el PATCH hecho con curl desde ese origen se ejecutó igual: 200 y la inscripción 3 quedó `confirmada` (`cors_curl_ejecuta.txt`). CORS no es autenticación: lo aplica el navegador y curl o Bruno lo ignoran. Lo que protege el cambio de estado son el token y el rol. Cuando publiquemos el front, hay que agregar su origen exacto (con `https`, el dominio y sin barra final), mejor desde una variable de entorno que escrito en el código.

### D3 · El límite del login
El tope es de 5 intentos por minuto por IP, con la ventana deslizante de `httprate`. Una persona que se equivoca de clave lo hace una, dos o tres veces. Una vuelta de la colección de Bruno inicia sesión tres veces y entra en el tope. En cambio, sin límite, cada intento cuesta unos 65 ms de bcrypt, y un atacante podía probar cientos de claves por minuto. Con el tope, una IP prueba como mucho 5 por minuto.

Va solo en `POST /login`, con `r.With`, porque es la única ruta donde se adivinan contraseñas y la más cara (bcrypt). Con `r.Use` frenaría también las rutas normales, y las seis peticiones chocarían con el tope. Como la ruta está en `internal/eventosports/rutas.go`, el límite llega como un campo `LimiteLogin` del manejador: `main.go` pone el de `httprate` y las pruebas pueden armar el router sin él.

En la ráfaga (`rafaga.txt`), los cinco primeros intentos dieron 401 con `X-Ratelimit-Remaining` de 4 a 0. El primer 429 llegó en el sexto, con `Retry-After: 60`, y el séptimo dio lo mismo. En el registro, cada 401 tardó entre 64 y 79 ms y cada 429 salió en `0s`: lo responde el middleware sin llegar a bcrypt. La prueba `internal/eventosports/limite_test.go` comprueba lo mismo, y además que `GET /eventos` sigue respondiendo con el login frenado.

### D4 · Detrás de un proxy
`ClientIPFromRemoteAddr` saca la IP de `r.RemoteAddr`, la dirección de la conexión TCP (sin el puerto), y no lee ninguna cabecera. Por eso en el registro aparece `from ::1` en vez de `[::1]:60358`. Si la API queda detrás de un proxy (nginx, el balanceador de la nube), todas las conexiones llegan desde la IP del proxy. Todos los usuarios compartirían una sola cuenta, y cinco intentos fallidos de cualquiera bloquearían el login de todos durante un minuto.

No se puede confiar en `X-Forwarded-For` sin más porque la escribe el cliente: un atacante pondría un valor distinto en cada petición y nunca llegaría al tope. Lo correcto es aceptarla solo cuando la conexión viene del proxy que conocemos, y tomar la IP que agregó ese proxy. Chi 5.5.0 lo trae: `middleware.ClientIPFromXFF("rango/del/proxy")`, o `ClientIPFromXFFTrustedProxies(1)` si hay exactamente un proxy delante y la red garantiza que nadie más llega a la API.

### Uso de IA en la semana 12
Claude Code (Claude Opus 5.5) diagnosticó la conexión a PostgreSQL; corrió las búsquedas, las pruebas, la ráfaga y las peticiones de Bruno; escribió el código de CORS, del límite, del usuario de la base (reto) y de las pruebas nuevas; y redactó la auditoría y D1 a D7 a partir de esas salidas. Carlos aprobó cada comando. Las evidencias de `docs/taller12/` son las salidas tal como las dieron los comandos. La pareja tiene que revisar estas decisiones para poder defenderlas sin IA en el hito 3.

## Semana 12 · reto: mínimo privilegio en la base

### D5 · Con qué usuario se conecta la API
Con `eventosports_api` (el de `DATABASE_URL`). No es superusuario, no puede crear bases ni usuarios, no hereda permisos y no es dueño de ninguna tabla (`reto_usuarios.txt`). Su clave es aleatoria y vive solo en el `.env` de cada máquina. En el repositorio solo está el marcador de `.env.example`. Los permisos, tabla por tabla (`reto_permisos.txt`), son exactamente los que usa la API:

| Tabla | Permisos | Por qué |
| --- | --- | --- |
| `cuentas` | `SELECT` | El login lee nombre, rol y resumen de la clave; la siembra la hace el dueño |
| `usuarios`, `sedes` | `SELECT` | Perfiles, filtro `?ciudad=` y vista del evento; la API no los cambia |
| `eventos` | `SELECT`, `INSERT`, `UPDATE (cupos)` | El organizador crea eventos, pero la API nunca los cambia. `UPDATE (cupos)` está solo porque PostgreSQL pide UPDATE en alguna columna para `SELECT … FOR UPDATE`, el bloqueo que evita vender dos veces el último cupo |
| `inscripciones` | `SELECT`, `INSERT`, `UPDATE (estado)` | Reservar y cambiar el estado; ninguna otra columna se cambia |
| `pagos` | `SELECT`, `INSERT`, `UPDATE (verificado)` | Registrar el comprobante y verificarlo |
| `eventos_id_seq`, `inscripciones_id_seq`, `pagos_id_seq` | `USAGE` | Los `id` nuevos salen de la secuencia, que tiene su propio permiso |

Ninguna tabla tiene `DELETE`, `TRUNCATE`, `REFERENCES` ni `TRIGGER`. Al intentar borrar una tabla, PostgreSQL respondió `BEGIN`, `ERROR:  must be owner of table inscripciones` y `ROLLBACK` (`reto_drop.txt`). En `reto_otras_ordenes.txt`, `DELETE FROM inscripciones` y `UPDATE eventos SET titulo` dieron `permission denied`, `CREATE TABLE` dio `permission denied for schema public` y `CREATE ROLE` también se negó. Con este usuario, las seis peticiones siguen en 401, 200, 200, 403, 200 y 422.

### D6 · Quién crea y cambia las tablas
Las crea y las cambia `eventosports`, el usuario dueño que crea el contenedor de Compose (`MIGRACION_DATABASE_URL`). Eso pasa con `go run . -migrar`: se conecta como dueño, corre `AutoMigrate` y el índice único, siembra a Ana y Luis si `DEMO_SEED=true` y llama a `datos.PrepararUsuarioAPI`. Esa función crea `eventosports_api` (o le actualiza la clave), le quita todos los permisos y le vuelve a dar solo los de la lista. Si alguien agregó un permiso a mano, la siguiente migración lo borra. `-migrar` también se niega a seguir si `DATABASE_URL` usa al mismo dueño. Las pruebas de integración usan `MIGRACION_DATABASE_URL`, porque crean y borran su propio esquema.

`go run .` ya no corre `AutoMigrate`. Si cambiamos un modelo y arrancamos la API con su usuario sin migrar, la API arranca igual, pero las consultas que usen la columna nueva fallarían con 500 (la columna no existe) hasta que el dueño corra `go run . -migrar`. Esto lo razonamos, no lo probamos. Y si alguien volviera a poner `AutoMigrate` con el usuario de la API, PostgreSQL lo frenaría: crear una tabla da `permission denied for schema public`, y cambiar una existente pide ser dueño, como el `DROP` de `reto_drop.txt`. El orden desde cero es: `docker compose up -d --wait db`, `go run . -migrar` y `go run .`. Se probó primero sobre la base que ya teníamos y después desde un volumen vacío, con `docker compose down -v` y `docker compose up -d --wait db` (`reto_desde_cero.txt`). Sin migrar, `go run .` no pudo conectarse, porque `eventosports_api` todavía no existía. `go run . -migrar` creó las seis tablas, todas del dueño, y el usuario de la API. Después `go run .` arrancó, las seis peticiones pasaron y el `DROP` volvió a dar `must be owner of table inscripciones`. En esta rama Compose solo levanta la base (la API con Compose era del taller 11).

### D7 · Lo que esto no frena
Si quedara una inyección, quien la encuentre tendría los permisos de `eventosports_api`. Podría leer todo lo que la API lee, incluidos los resúmenes bcrypt de `cuentas` y los datos de los usuarios. También podría crear eventos e inscripciones, cambiar el estado de cualquier inscripción, marcar pagos como verificados y cambiar los `cupos` de un evento. Lo que ya no puede: borrar tablas o registros, vaciar tablas, crear tablas o usuarios, cambiar otras columnas, darse más permisos ni tocar la configuración del servidor. Eso lo podía hacer antes, cuando la API se conectaba como `eventosports`, que es superusuario. Tampoco cubre lo que PostgreSQL da a todos por defecto (`PUBLIC`), como conectarse a otras bases del mismo servidor o crear tablas temporales. La defensa principal sigue siendo la consulta con parámetros del punto 1; el mínimo privilegio limita el daño si esa falla.
