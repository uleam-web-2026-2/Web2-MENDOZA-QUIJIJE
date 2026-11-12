# Hito 2 · EventoSports · Grupo E · Web II A

**Sebastián Mendoza · Evelin Quijije** · Entidad principal: Inscripción. Antes de cada turno: ejecutar petición **0** en Bruno; prepara registros nuevos y variables de ejecución. Entorno **Local**, `http://localhost:8080`.

## 1 · Usuarios de ejemplo

| Usuario | Rol | Contraseña de ejemplo |
|---|---|---|
| ana | jugador | ana-2026 |
| luis | organizador | luis-2026 |

## 2 · Las seis peticiones

| # | Petición | Credencial | Debe responder |
|---|---|---|---|
| 1 | `GET /eventos/{{eventoId}}/inscripciones` | Ninguna | 401 |
| 2 | `POST /login` con Luis | Ninguna | 200 y token |
| 3 | La misma ruta de 1 | Token de Luis | 200 |
| 4 | `PATCH /inscripciones/{{inscripcionId}}/estado` → confirmada | Token de Ana | 403 |
| 5 | La misma petición de 4 | Token de Luis | 200 |
| 6 | `PATCH /inscripciones/{{canceladaId}}/estado` → confirmada | Token de Luis | 422 |

En 4–5 la inscripción está en `pago_por_verificar`: solo el organizador dueño puede verificar. En 6 está `cancelada`, estado final. Los tokens e IDs los prepara la colección; no se pega un token de otro día.

## 3 · Matriz ruta × rol

| Ruta implementada | Sin sesión | Jugador | Organizador |
|---|---|---|---|
| `POST /login` | Sí, con credenciales | Sí | Sí |
| `GET /eventos` | Sí | Sí | Sí |
| `GET /eventos/{id}` | Sí | Sí | Sí |
| `POST /eventos` | 401 | 403 | Sí |
| `POST /eventos/{id}/inscripciones` | 401 | Sí, para sí mismo | 403 |
| `GET /eventos/{id}/inscripciones` | 401 | 403 | Solo sus eventos |
| `GET /inscripciones/{id}` | 401 | Solo su inscripción | Solo sus eventos |
| `PATCH /inscripciones/{id}/estado` | 401 | Propia y transición permitida | Evento propio y transición permitida |

## 4 · Recuadro de cumplimiento

☑ Structs, relaciones y tablas · ☑ Contrato sin errores ni avisos · ☑ Peticiones 0–6 en orden · ☑ Respuestas coinciden con esta hoja · ☑ Pruebas en verde e integración PostgreSQL · ☑ Capas explicadas en `docs/decisiones.md` · ☑ Capturas del informe de ejecución verificadas.
