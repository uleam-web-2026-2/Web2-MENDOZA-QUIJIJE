# Semana 13 · las seis peticiones del hito 2

| Petición | Antes (taller 12) | Con el WebSocket | Con el reto |
| --- | --- | --- | --- |
| 1 sin token | 401 |  |  |
| 2 login | 200 |  |  |
| 3 con token | 200 |  |  |
| 4 rol sin permiso | 403 |  |  |
| 5 cambio permitido | 200 |  |  |
| 6 transición prohibida | 422 |  |  |

La columna «Con el reto» se llena solo si hicieron el reto.
Antes de cada vuelta, el servidor se arrancó otra vez: el límite del taller 12
cuenta los inicios de sesión durante un minuto.

**Antes (taller 12).** Corrida el 27/11/2026 con `go run .` en el puerto 8080 (rama `taller13`,
recién instalada `coder/websocket`, todavía sin usar) y la colección `docs/hito2_bruno` desde
Bruno CLI 4.3.0 (`npx --yes @usebruno/cli@4.3.0 run --env Local --bail`). La petición 0 crea un
evento y una inscripción nuevos, así que la 5 no gasta un registro viejo. Salida completa:
[`bruno_antes.txt`](bruno_antes.txt).
