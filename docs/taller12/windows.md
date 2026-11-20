# Taller 12 en Windows: qué falló y qué hicimos

Equipo: Windows 11, Go 1.27.1, Docker Desktop, Git Bash para todos los comandos y evidencias
(así `>` guarda en UTF-8 y `curl` es el curl real).

| Qué falló | Por qué | Qué hicimos |
| --- | --- | --- |
| `go run .` no conectaba a PostgreSQL | En `servidor/.env`, `DATABASE_URL` apuntaba a `sqlserver://localhost:1433` (de otra materia), no a la base del proyecto | Se dejó `DATABASE_URL=postgres://eventosports:…@localhost:5433/eventosports?sslmode=disable`, con la misma clave de `POSTGRES_PASSWORD` |
| `docker` respondía `open //./pipe/dockerDesktopLinuxEngine: The system cannot find the file specified.` | Docker Desktop estaba cerrado | Se abrió Docker Desktop y después `docker compose up -d db` en `servidor/` |
| `go test ./...`: `no se pudo conectar a la base de pruebas`, y por red `password authentication failed for user "eventosports"` | El volumen `eventosports-s7_postgres_data` se creó con una clave anterior; `POSTGRES_PASSWORD` solo se usa la primera vez que se crea el volumen | Sin borrar el volumen: `ALTER ROLE eventosports PASSWORD ...` por el socket local del contenedor, con la clave del `.env` |
| Confusión de puertos | En la máquina también corre el servicio `postgresql-x64-16` de Windows en el 5432 | No se tocó: el proyecto usa el 5433 publicado por Compose |

Dentro del contenedor, las conexiones por `127.0.0.1` entran sin clave (`trust` en `pg_hba.conf`): para
probar la contraseña de verdad hay que conectar por la IP de la red de Docker o desde Windows.
