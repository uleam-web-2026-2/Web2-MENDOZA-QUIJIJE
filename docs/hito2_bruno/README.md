# Colección de Bruno · Hito 2

Contiene **siete peticiones**, numeradas del 0 al 6. Se probó con Bruno CLI 4.3.0. `bruno.json` es la raíz de la colección; el entorno `Local` solo contiene la dirección y un ID inicial de referencia, sin tokens guardados.

## En el cliente de Bruno

1. Abrir esta carpeta como colección y seleccionar **Local**.
2. Con la base y el servidor encendidos, enviar **0 - Preparar registros nuevos** antes del turno.
3. Comprobar que la preparación responde 201 y que no hay errores de scripts. Crea una inscripción nueva y deja un comprobante pendiente mediante un script posterior. También deja preparada otra inscripción cancelada.
4. Sin cerrar la colección, enviar **1, 2, 3, 4, 5 y 6**, en ese orden. No hay que editar archivos ni pegar tokens.

La 0 guarda los IDs y el token de Ana en variables de ejecución. La 2 obtiene el token de Luis en vivo. Se usa `bru.setVar`, que conserva variables durante la ejecución; no se usa `setEnvVar` para guardar tokens en disco.

La 5 consume el estado de la inscripción: repetirla devuelve 422. Para ensayar de nuevo, volver a enviar la 0. La 6 no modifica su registro y puede repetirse.

## Desde la terminal

Desde la raíz del repositorio:

```powershell
cd docs/hito2_bruno
npx --yes @usebruno/cli@4.3.0 run --env Local --bail
```

Si la API está en otro puerto, se puede sobrescribir la dirección para esa ejecución, por ejemplo:

```powershell
npx --yes @usebruno/cli@4.3.0 run --env Local --env-var baseUrl=http://localhost:18080 --bail
```

El recorrido completo incluye la preparación. Para demostrar petición por petición se puede usar el cliente de Bruno. Si cambian el puerto, ajustar `baseUrl` en el entorno **antes** del turno.

No ejecutar con `--parallel`: los pasos dependen de las variables y estados de los anteriores. La preparación usa `bru.sendRequest`, disponible en el modo seguro del CLI probado; no requiere acceso al disco desde los scripts.

Referencia de las APIs usadas: [JavaScript de Bruno](https://docs.usebruno.com/testing/script/javascript-reference).
