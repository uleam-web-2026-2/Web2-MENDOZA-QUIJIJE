# Mesa de ayuda — servidor base (Semana 2, Taller Chi)

Servidor HTTP en Go con [chi](https://github.com/go-chi/chi) para la mesa de ayuda de
tickets. Sin base de datos todavía: los tickets viven en memoria y se pierden al apagar
el servidor.

## Cómo correrlo

```bash
go run .
```

Levanta en `:8080`. Prueba todo el contrato con:

```bash
./pruebas.sh
# o contra otra máquina:
./pruebas.sh http://IP:8080
```

## Tabla de rutas

| Método | Ruta            | Éxito              | Errores posibles                              |
|--------|-----------------|--------------------|------------------------------------------------|
| GET    | /               | 200 info del servicio | —                                             |
| GET    | /tickets        | 200 lista          | —                                                |
| POST   | /tickets        | 201 ticket creado  | 400 cuerpo_malformado · 422 datos_invalidos     |
| GET    | /tickets/{id}   | 200 ticket         | 400 id_invalido · 404 no_encontrado             |
| GET    | /explotar       | —                  | 500 error_interno                               |
| *      | ruta inexistente| —                  | 404 ruta_inexistente                            |
| *      | método no soportado | —              | 405 metodo_no_permitido                         |

Toda respuesta sigue la misma envoltura:

```json
{ "ok": true, "datos": { ... } }
{ "ok": false, "error": { "codigo": "...", "mensaje": "..." } }
```

## Decisiones de diseño

Ver [`docs/decisiones.md`](docs/decisiones.md).

## Revisión cruzada

Revisado por: _pendiente_ — sin hallazgos.