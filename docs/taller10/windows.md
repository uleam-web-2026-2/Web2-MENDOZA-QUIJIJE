# Semana 10 · Problemas resueltos en Windows

## Bruno ejecutado desde otra carpeta

El CLI de Bruno respondió que solo podía ejecutarse desde la raíz de una colección cuando se invocó desde la raíz del repositorio. La colección de este proyecto está en `docs/hito2_bruno`, donde se encuentra `bruno.json`. Se cambió a esa carpeta antes de ejecutar `npx --yes @usebruno/cli@4.3.0 run --env Local --bail`.

En otra prueba, `go run .` se ejecutó desde `docs/hito2_bruno` después de intentar entrar en una carpeta `servidor` que no existe dentro de ella. Desde `docs/hito2_bruno`, la ruta correcta es `../../servidor`. Bruno también dio un error de conexión cuando la API todavía no escuchaba en `localhost:8080`; se inició el servidor antes de repetir la colección.

## Salida de Redocly en PowerShell

Al unir la salida normal y la de errores con `2>&1`, PowerShell convirtió el aviso de actualización de Redocly en objetos `RemoteException` y mostró caracteres mal codificados. Además, el primer intento se hizo desde `servidor/`, donde no existe `docs/taller10/`. Desde la raíz del repositorio, se guardó la salida sin modificarla mediante `cmd /c` y se desactivó el aviso de actualización solo para esa ejecución:

```powershell
cmd /c "set CI=1&& npx --yes @redocly/cli@2.54.3 lint docs/openapi.yaml > docs\taller10\redocly.txt 2>&1"
```

El archivo resultante es texto UTF-8 legible y Redocly confirma que `docs/openapi.yaml` es válido.
