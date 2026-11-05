# Peticiones del hito 2: antes y después

| Petición | Antes | Después |
| --- | --- | --- |
| 1 sin token | 401 | |
| 2 login | 200 | |
| 3 con token | 200 | |
| 4 rol sin permiso | 403 | |
| 5 cambio permitido | 200 | |
| 6 transición prohibida | 422 | |

**De dónde sale «Antes».** Es lo que respondió el servidor el 5 de noviembre de 2026, antes de cambiar el código de `servidor/` para este taller, con PostgreSQL 16 instalado en el equipo. Desde `docs/hito2_bruno`:

```bash
npx --yes @usebruno/cli@4.3.0 run --env Local --bail
```

Resultado: 7 de 7 peticiones y 9 de 9 comprobaciones, contando la petición 0 de preparación, que respondió 201.
