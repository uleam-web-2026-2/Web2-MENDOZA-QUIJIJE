# Semana 11 · problemas resueltos en Windows

Se trabajó con Windows 11, PowerShell 5.1 y Docker Desktop 4.84 (motor 29.6.2, WSL 2).

## 1. Docker no respondía

**Qué falló.** `docker info` y los demás comandos respondían:

```text
failed to connect to the docker API at npipe:////./pipe/dockerDesktopLinuxEngine;
... open //./pipe/dockerDesktopLinuxEngine: The system cannot find the file specified.
```

**Por qué.** El programa `docker` estaba instalado, pero Docker Desktop no estaba abierto, así que no había motor al otro lado. En este equipo Docker Desktop está instalado para el usuario, en `%LOCALAPPDATA%\Programs\DockerDesktop`, y no en `Program Files`.

**Qué hicimos.** Abrir Docker Desktop, esperar a que dijera que estaba corriendo y comprobar con `docker run --rm hello-world`.

## 2. Las evidencias aparecían como binarias en GitHub

**Qué falló.** El primer `imagenes.txt` y el primer `capas.txt` se veían en GitHub como archivos binarios.

**Por qué.** En PowerShell 5.1, `>` guarda el texto en UTF-16, y GitHub no lo muestra como texto. Los archivos empezaban con los bytes `FF FE`.

**Qué hicimos.** Generar cada evidencia a través de `cmd`, que deja pasar los bytes tal como los escribe Docker (UTF-8), desde `servidor/`:

```bash
cmd /c "docker image ls eventosports > ..\docs\taller11\imagenes.txt"
```

Los demás archivos de `docs/taller11/` se generaron igual.
