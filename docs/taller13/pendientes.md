# Taller 13 · pendientes

## Punto 3 · la captura de las dos ventanas y `logs_avisos.txt`

**Qué probamos.** A las 22:13 arrancamos la API (`go run .`, puerto 8080) y la página del paquete
(`go run servir.go`, en `http://localhost:5173`). A las 22:14 creamos con curl un evento pagado de
luis (31), la inscripción 5 con el token de ana, y le cambiamos el estado dos veces:
ana a `pago_por_verificar` y luis a `confirmada`. Las tres respuestas fueron 201, 200 y 200.

**Qué salió.** En el registro del servidor no apareció ningún `POST /login` de la página ni ninguna
línea `GET /ws`, y `netstat` no mostró conexiones de un navegador al 5173 ni al 8080: las ventanas
de ana y eva no llegaron a conectarse a esa API, así que esos dos avisos no tuvieron a quién llegar.
Por eso todavía no hay `aviso_navegador.png` ni `logs_avisos.txt` de una misma vuelta.

**Lo que sí está comprobado sin navegador.** `TestElAvisoLlegaSoloAlJugadorDueno`
(`prueba_websocket.txt`) conecta a ana y a eva por WebSocket a un servidor real: a ana le llega el
aviso y a eva no. Los `curl` del paso 2.3 dieron 401, 403 y 101.

**Segunda vuelta, 23:05.** Con la API del reto arrancada con `go run .` y la misma página, abrimos
`http://localhost:5173` en dos pestañas del navegador integrado de la app de Claude (un navegador
real): ana a las 23:05:04 y eva a las 23:05:14, las dos con «Conectado como …». Con curl, luis creó
el evento 36, ana se inscribió con su token (inscripción 10), ana la pasó a `pago_por_verificar` y
luis a `confirmada` (201, 200, 200). En la pestaña de ana aparecieron, sin recargar,
`estado_cambiado · registro 10 → pago_por_verificar` y `→ confirmada`. En la de eva, nada. Ese panel
muestra una pestaña a la vez, así que no permite la captura de dos ventanas juntas. Falta guardar la
captura y copiar el registro del servidor de esta vuelta a `logs_avisos.txt`.
