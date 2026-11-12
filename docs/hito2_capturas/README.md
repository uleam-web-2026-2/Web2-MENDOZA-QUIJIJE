# Respaldos de la ejecución real

Estas imágenes muestran el **informe de respuestas reales de Bruno CLI 4.3.0**, ejecutado contra EventoSports en `localhost:8080`, después de separar las capas. No son imágenes de la interfaz del cliente de Bruno ni respuestas inventadas. El texto e imágenes se exportaron del reporte JSON de la ejecución.

| Imagen | Evidencia |
|---|---|
| `0_0_preparar.png` | Creación 201 de la inscripción del ensayo |
| `1_1_sin_token.png` | Ruta protegida sin sesión: 401 |
| `2_2_login.png` | Login: 200 y campo token, oculto al exportar |
| `3_3_con_token.png` | Misma ruta con sesión: 200 |
| `4_4_rol_sin_permiso.png` | Jugador intenta verificar pago: 403 |
| `5_5_rol_correcto.png` | Organizador verifica: 200 |
| `6_6_transicion_prohibida.png` | Cancelada no vuelve a confirmada: 422 |
| `7_pruebas.png` | Salida real de `go test ./... -count=1` |
| `8_contrato.png` | Salida real del validador, sin errores ni avisos |

`resultados_bruno.json` conserva el reporte exportado, códigos, cuerpos y resultados de las comprobaciones. Se ocultaron los valores de `Authorization` y `token`; no se guardó el secreto de firma. Las credenciales de ejemplo son las del taller y no son contraseñas personales.

La respuesta de la petición 0 contiene el estado inicial `reservada`, porque esa es la respuesta de crear. Su script posterior envía el comprobante para dejar el registro en `pago_por_verificar` antes de la petición 4. La captura 3 muestra el comprobante ya preparado. No se alteró la respuesta 0 para aparentar otro resultado.

Las capturas son el respaldo de esa ejecución, con los IDs que existían entonces. La colección usa IDs nuevos al volver a ejecutar la 0. Para la demostración pueden complementar estos informes con capturas directas de su cliente Bruno, especialmente si el docente las prefiere.

No se ha marcado como realizado el ensayo oral ni la entrega en Git: esas acciones corresponden a la pareja.
