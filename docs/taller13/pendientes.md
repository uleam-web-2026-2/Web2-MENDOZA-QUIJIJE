# Taller 13 · pendientes y cómo se hizo el punto 3

## La captura y el registro (vuelta de las 23:43)

Con la API del reto arrancada con `go run .` (puerto 8080) y la página del paquete con
`go run servir.go` (puerto 5173), la abrimos en el navegador integrado de la app de Claude, que es un
navegador real. Ese panel muestra una sola pestaña a la vez. Por eso pusimos **dos copias de la
página real** (`http://localhost:5173`) una al lado de la otra, cada una en un `iframe`, en una
página auxiliar servida en el 5174. Cada copia inicia sesión y abre su propio WebSocket con el
origen `http://localhost:5173`. La captura muestra un solo navegador con las dos páginas lado a lado,
**no dos ventanas separadas**: así lo pudimos hacer, y lo dejamos dicho.

- 23:43:01: ana entra a la izquierda y eva a la derecha. Las dos dicen «Conectado como …».
- 23:43:10: luis crea el evento 37 con curl, y ana se inscribe con su token (inscripción 11).
- 23:43:12: ana pasa la inscripción a `pago_por_verificar`. 23:43:14: luis la pasa a `confirmada`.
- En la página de ana aparecen, sin recargar, `estado_cambiado · registro 11 → pago_por_verificar` y
  `→ confirmada`. En la de eva, ningún aviso.

[`aviso_navegador.png`](aviso_navegador.png) es la captura que tomó la herramienta del navegador
integrado. Solo la pasamos de JPG a PNG, sin recortes ni retoques. [`logs_avisos.txt`](logs_avisos.txt)
tiene las líneas del servidor de esa misma vuelta, con el token borrado con el `sed` del taller. Los
dos primeros `POST /login` son las páginas de ana y eva, y los otros dos son luis y ana desde curl.
Las dos líneas `GET /ws?token=BORRADO … 101` (36,7 s) aparecen al cerrar el panel.

## Antes: vueltas que no sirvieron para la captura

- **22:14.** Hicimos los cambios de estado de la inscripción 5, pero ninguna ventana del navegador se
  había conectado a la API: en el registro no hubo ningún `POST /login` de la página.
- **23:05.** Ana y eva se conectaron en dos pestañas del navegador integrado, y ana recibió los dos
  avisos de la inscripción 10. Pero el panel se cerró a las 23:05:59, antes de guardar la captura.
- Desde Brave, en esta computadora, el servidor nunca recibió un inicio de sesión de la página.
