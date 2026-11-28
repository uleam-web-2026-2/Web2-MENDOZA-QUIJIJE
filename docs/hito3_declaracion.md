# Hito 3 · Declaración de EventoSports

Pareja: Carlos Sebastián Mendoza Avellán y Evelin Quijije Bailon · Aplicaciones Web II · 2026-2

## Plataforma del cliente y por qué

**Primero web (Vite, en `http://localhost:5173`), después una app Android para los jugadores.**

En EventoSports hay dos usuarios con necesidades distintas. El **organizador** publica eventos, revisa
la lista de inscritos y verifica los comprobantes de pago: lo hace sentado, con pantalla grande, y le
sirve la web. El **jugador** reserva su cupo y paga por transferencia desde la app de su banco, en el
celular, y a veces con señal irregular (las canchas de Manta, como dice la ficha del hito 1). Por eso,
después de la web, el cliente natural del jugador es una app Android.

Empezamos por la web porque la API ya está preparada para ese origen (CORS y `/ws` aceptan solo
`origenFront`) y se prueba en cualquier navegador. Por pensar en Android, los avisos van por
**WebSocket** y no por SSE: `EventSource` es del navegador, mientras que la misma ruta `/ws?token=…`
sirve en el navegador y en Android.

## Contrato de los mensajes de pagos

Los mensajes llegan por `GET /ws?token=…`, como **un objeto JSON por mensaje**, con los mismos
campos base que ya usamos (`tipo`, `id` de la inscripción y `estado` después del cambio). Hoy los
emite `negocio.Servicio.CambiarEstado`; en el hito 3 los emitirá el servicio de pagos con este mismo
contrato.

| `tipo` | Cuándo se envía | Quién lo recibe | Campos además de `tipo`, `id` y `estado` |
| --- | --- | --- | --- |
| `pago_registrado` | El jugador sube el comprobante: `reservada` → `pago_por_verificar` | Solo el organizador **del evento** | `evento_id`, `pago_id`, `monto_centavos`, `banco`, `referencia` |
| `pago_verificado` | El organizador lo verifica: `pago_por_verificar` → `confirmada` | Solo el jugador dueño de la inscripción | `pago_id` |
| `pago_rechazado` | El organizador lo devuelve: `pago_por_verificar` → `reservada` | Solo el jugador dueño de la inscripción | `pago_id` |

Ejemplo: `{"tipo":"pago_verificado","id":11,"estado":"confirmada","pago_id":7}`

Reglas del contrato:
- Se envía **después** de guardar (después del COMMIT): nunca se avisa de un pago que no quedó en la base.
- El monto va en **centavos enteros** (`250` = 2,50 USD), para no depender de decimales con `float`.
- Un aviso puede perderse si la conexión está lenta o cerrada. El dato correcto siempre está en
  `GET /inscripciones/{id}`, y `pago_id` permite descartar un aviso repetido.
- El cliente ignora los `tipo` que no conoce, así podemos agregar mensajes sin romper versiones viejas de la app.
- Nunca viajan el token, la contraseña ni datos de la cuenta del jugador. El nombre del jugador
  tampoco: el organizador lo ve en `GET /eventos/{id}/inscripciones`, que revisa que el evento sea suyo.
