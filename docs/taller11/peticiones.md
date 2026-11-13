# Semana 11 · las seis peticiones del hito 2

| Petición | Con go run | En el contenedor | Con Compose |
| --- | --- | --- | --- |
| 1 sin token | 401 | 401 | 401 |
| 2 login | 200 | 200 | 200 |
| 3 con token | 200 | 200 | 200 |
| 4 rol sin permiso | 403 | 403 | 403 |
| 5 cambio permitido | 200 | 200 | 200 |
| 6 transición prohibida | 422 | 422 | 422 |

Ningún código cambió. Las tres columnas se tomaron el 12 de noviembre de 2026 con Bruno CLI 4.3.0, desde `docs/hito2_bruno`, contra `http://localhost:8080`:

```bash
npx --yes @usebruno/cli@4.3.0 run --env Local --bail
```

En las tres ejecuciones, la petición 0 preparó registros nuevos (201), y pasaron 7 de 7 peticiones y 9 de 9 comprobaciones.

- **Con go run:** `go run .` en `servidor/`, con PostgreSQL 16 instalado en Windows, en `localhost:5432`.
- **En el contenedor:** `docker run -d --name eventosports-api -p 8080:8080 --env-file .env.docker eventosports`, con la misma base a través de `host.docker.internal:5432`. El log con las seis peticiones está en `logs_contenedor.txt`.
- **Con Compose:** `docker compose up -d --build` en `servidor/`, con la base de Compose. Ver `compose.txt`.

## Petición 3 con go run (respuesta completa)

`GET /eventos/49/inscripciones`:

```json
[{"id":24,"evento_id":49,"jugador_id":1,"estado":"pago_por_verificar","creado":"2026-11-12T22:26:53.4017-05:00","jugador":{"id":1,"nombre":"ana"},"ultimo_pago":{"id":11,"inscripcion_id":24,"monto":2.5,"banco":"Banco de ejemplo","referencia":"HITO2-1794540413574","verificado":false,"creado":"2026-11-12T22:26:53.58076-05:00"}}]
```

## La misma, en el contenedor (paso 2.5)

`GET /eventos/51/inscripciones`:

```json
[{"id":26,"evento_id":51,"jugador_id":1,"estado":"pago_por_verificar","creado":"2026-11-13T03:29:48.159934Z","jugador":{"id":1,"nombre":"ana"},"ultimo_pago":{"id":12,"inscripcion_id":26,"monto":2.5,"banco":"Banco de ejemplo","referencia":"HITO2-1794540588270","verificado":false,"creado":"2026-11-13T03:29:48.282797Z"}}]
```

Los IDs cambian porque la petición 0 crea un evento y una inscripción nuevos en cada ejecución. Lo que importa es el formato de `creado`: con `go run` lleva `-05:00`, y en el contenedor, `Z`. Ver D4 en `docs/decisiones.md`.
