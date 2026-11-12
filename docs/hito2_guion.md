# Hito 2 · Guion para exponer individualmente

**EventoSports · Grupo E · Aplicaciones Web II A**

El enunciado oficial establece una demostración en vivo de cinco minutos y especifica que no hay diapositivas. Este guion sirve si te toca demostrar sin tu pareja: debes realizar las cinco demostraciones y responder las dos preguntas. Practica también las seis preguntas posibles; el docente elige cuáles hacer.

## Preparación

- Inicia PostgreSQL y la API antes de que llamen a tu pareja.
- En Bruno, selecciona **Local** y envía `0 - Preparar registros nuevos`. Luego deja abierta la petición 1.
- Deja abierta una terminal en la raíz para `git status --short`, `git log -1` y el validador Redocly.
- Deja lista otra terminal en `servidor/` para `go test ./...`.
- Abre `servidor/internal/eventosports/estado.go`, `servidor/internal/negocio/estados.go` y `docs/decisiones.md`.
- Si el puerto 8080 está ocupado, arranca el servidor en uno libre, por ejemplo `$env:PORT='18080'; go run .`, y usa ese mismo puerto en la colección.
- Ten preparadas las capturas de respaldo en `docs/hito2_capturas/`. No muestres ni copies el valor del JWT.

## Demostración, minuto a minuto

| Tiempo | Qué haces |
|---|---|
| 0:00–0:25 | Muestras `git status --short` y `git log -1`. La entrega debe estar subida; el estado debe estar limpio salvo el ID de prueba permitido por el enunciado. |
| 0:25–1:00 | Ejecutas `npx @redocly/cli lint docs/openapi.yaml --format=stylish`. Señalas `security` y las respuestas 401, 403 y 422 en el contrato. |
| 1:00–1:50 | Envías Bruno 1, 2 y 3: sin token, login de Luis, y la misma consulta con el token. |
| 1:50–2:35 | Envías Bruno 4 y 5: Ana no puede verificar; Luis, organizador dueño, sí puede. |
| 2:35–3:20 | Envías Bruno 6 para ver el 422 y ejecutas `go test ./...`. |
| 3:20–3:50 | Muestras las carpetas y los dos archivos de capas; lo comparas con `docs/decisiones.md`. |
| 3:50–4:50 | Respondes dos preguntas del docente. Como expones solo, respondes ambas tú. |
| 4:50–5:00 | El docente registra la nota. |

## Frases breves para cada demostración

### 1 · El contrato

> Este es el contrato de la API. El validador termina sin errores ni avisos. Aquí están declarados `security` y las respuestas 401, 403 y 422.

### 2 · Quién es

> Sin token, la consulta de inscripciones responde 401. Inicio sesión como Luis y recibo un token nuevo. Con ese token, la misma consulta responde 200.

### 3 · Qué puede hacer

> Ana es jugadora y no puede verificar el pago: recibe 403. Luis es el organizador dueño del evento y sí puede verificarlo: recibe 200.

### 4 · La regla del negocio

> Una inscripción cancelada no puede volver a confirmarse. La API responde 422 y las pruebas terminan en verde.

### 5 · La estructura

> El manejador HTTP recibe la solicitud y delega el cambio. La regla de negocio decide si la transición es válida sin depender de HTTP. El acceso a PostgreSQL está en la capa de datos; esta estructura coincide con `docs/decisiones.md`.

## Las seis preguntas

### 1. ¿En qué línea se decide que la ruta pide token?

En `servidor/internal/eventosports/rutas.go`, dentro del grupo `r.Group`, se registra `r.Use(tokens.Autenticar)`. Las rutas de ese grupo pasan primero por el middleware. `SoloRol` agrega una restricción de rol a las operaciones que lo requieren.

### 2. ¿Qué lleva el token? ¿Qué pasa si alguien cambia el rol?

Incluye el ID (`sub`), nombre, rol, emisión (`iat`) y vencimiento (`exp`). Está firmado con HS256 y dura una hora. No está cifrado, pero cambiar el rol invalida la firma; el middleware rechaza el token y responde 401.

### 3. ¿Por qué la acción demostrada es solo del organizador?

La acción demostrada es verificar un pago. El organizador recibe y contrasta la transferencia; el jugador no debe confirmar su propio pago. Además, el negocio comprueba que el organizador sea dueño del evento. Otras transiciones del PATCH sí corresponden al jugador, por eso el permiso se decide por operación y propiedad.

### 4. Si alguien saca la ruta del grupo protegido, ¿qué prueba lo detecta?

`TestSinTokenLasRutasProtegidasResponden401`, en `servidor/internal/eventosports/sin_token_test.go`, envía cada operación protegida sin token y exige 401.

### 5. ¿Por qué 422 y no 400 o 403? ¿Dónde se valida?

La petición está bien formada y el usuario tiene acceso al evento, pero la transición desde `cancelada` está prohibida por la regla del negocio. Por eso no es 400 (sintaxis) ni 403 (permiso): corresponde 422. La regla está en `servidor/internal/negocio/estados.go`; la capa HTTP traduce el error en `servidor/internal/eventosports/manejadores.go`.

### 6. ¿Por qué eligieron esta estructura y qué cambiaría con otra base?

Separamos HTTP, reglas, persistencia y modelos para que una modificación de consultas no obligue a cambiar la regla de negocio. Si se cambia PostgreSQL, se reemplaza el adaptador en `servidor/internal/datos/` y se ajusta la conexión en `main.go`; las reglas de `negocio/` y sus pruebas puras pueden mantenerse.

## Alcance de equipos

El proyecto actual modela cinco entidades: Usuario, Sede, Evento, Inscripción y Pago. La inscripción corresponde a un jugador individual; no existe entidad Equipo ni administración de plantillas. Un club puede actuar como organizador, pero el sistema todavía no inscribe equipos completos. Para eso habría que ampliar el modelo y definir cómo se vinculan los equipos, sus jugadores y las inscripciones grupales.

## Lista final según el enunciado

- [ ] Structs, relaciones y tablas de todas las entidades del Hito 1.
- [ ] Contrato validado, con `security`, 401, 403, 422 y endpoints futuros marcados `(plan)`.
- [ ] Hoja de una página con sus cuatro secciones completas.
- [ ] Colección de Bruno con petición 0 y seis peticiones en orden.
- [ ] Pruebas en verde y al menos una prueba de integración ejecutada.
- [ ] Captura de cada respuesta y de las pruebas.
- [ ] Capa HTTP sin consulta directa a la base; regla de estado independiente de HTTP.
- [ ] Declaración de uso de IA en la entrega, como pide el enunciado.
- [ ] Commits reales de los dos integrantes y push al repositorio de la organización antes del cierre.

La exposición final no lleva PowerPoint: el enunciado pide demostrar la API, las pruebas y el código en vivo.
