# Taller 13 en Windows: qué falló y qué hicimos

Equipo: Windows 11, Go 1.27.1, Docker Desktop, Git Bash para los comandos y las evidencias
(así `>` guarda en UTF-8 y `curl` es el curl real, no el alias de PowerShell).

| Qué falló | Por qué | Qué hicimos |
| --- | --- | --- |
| `docker` respondía `open //./pipe/dockerDesktopLinuxEngine: The system cannot find the file specified.` | Docker Desktop estaba cerrado (igual que en el taller 12) | Se abrió Docker Desktop y después `docker compose up -d --wait db` en `servidor/` |
| `go test -race ./internal/eventosports -run 'WebSocket\|Aviso'` respondió `go: -race requires cgo; enable cgo by setting CGO_ENABLED=1` | En Windows, `-race` necesita un compilador de C (gcc) y la máquina no tiene | Como dice el anexo B: se siguió sin `-race`. `go test ./...` pasa; el mapa de `Avisos` se toca solo con `a.mu` tomado (`Suscribir`, `Quitar` y `Enviar`) |
