# Semana 12 · las seis peticiones del hito 2

| Petición | Antes (taller 11) | Con las defensas | Con el usuario de la API (reto) |
| --- | --- | --- | --- |
| 1 sin token | 401 | 401 | |
| 2 login | 200 | 200 | |
| 3 con token | 200 | 200 | |
| 4 rol sin permiso | 403 | 403 | |
| 5 cambio permitido | 200 | 200 | |
| 6 transición prohibida | 422 | 422 | |

**Antes (taller 11).** Corrida el 20/11/2026 con `go run .` en el puerto 8080 y la colección
`docs/hito2_bruno` desde Bruno CLI 4.3.0 (`npx --yes @usebruno/cli@4.3.0 run --env Local --bail`).
La petición 0 crea los registros nuevos, así que la 5 no gasta uno viejo. Salida completa:
[`bruno_antes.txt`](bruno_antes.txt).

**Con las defensas.** CORS, `ClientIPFromRemoteAddr`, el registro de peticiones y el límite de
5 por minuto en `POST /login`. Se reinició el servidor después de la ráfaga y se corrió la misma
colección: ningún código cambió. Cada vuelta hace tres inicios de sesión (dos en la petición 0 y la
petición 2), por debajo del tope. Salida: [`bruno_con_defensas.txt`](bruno_con_defensas.txt);
registro del servidor: [`logs_seis.txt`](logs_seis.txt).
