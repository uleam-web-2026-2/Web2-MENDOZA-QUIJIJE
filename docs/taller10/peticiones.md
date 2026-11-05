# Peticiones del hito 2: antes y después

| Petición | Antes | Después |
| --- | --- | --- |
| 1 sin token | 401 | 401 |
| 2 login | 200 | 200 |
| 3 con token | 200 | 200 |
| 4 rol sin permiso | 403 | 403 |
| 5 cambio permitido | 200 | 200 |
| 6 transición prohibida | 422 | 422 |

**De dónde sale «Antes».** Es lo que respondió el servidor el 5 de noviembre de 2026, antes de cambiar el código de `servidor/` para este taller, con PostgreSQL 16 instalado en el equipo. Desde `docs/hito2_bruno`:

```bash
npx --yes @usebruno/cli@4.3.0 run --env Local --bail
```

Resultado: 7 de 7 peticiones y 9 de 9 comprobaciones, contando la petición 0 de preparación, que respondió 201.

## Fase 4 · Comprobación después de conectar sqlc

Se repitió la colección de Bruno con el adaptador sqlc conectado. La petición 0
preparó datos nuevos y respondió 201. Las seis peticiones del hito 2 devolvieron
los mismos códigos de «Antes»: 401, 200, 200, 403, 200 y 422.

La salida de `go test ./...` se guardó en [go_test.txt](go_test.txt). Los
paquetes con pruebas muestran `ok` y no aparece `FAIL`.

Uso de IA: Codex ayudó a contrastar los códigos HTTP y a organizar la evidencia.