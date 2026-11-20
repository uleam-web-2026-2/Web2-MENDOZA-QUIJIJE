# Semana 12 · auditoría de inyección

Búsqueda (desde la raíz del repositorio): `git grep -n -E -f docs/taller12/patrones.txt -- "*.go" ":(exclude)*_test.go"`.
Salida completa en `auditoria.txt` (27 líneas). Todas las rutas empiezan en `servidor/internal/` salvo `servidor/main.go`.

| Línea | ¿Arma SQL? | ¿Llega texto del cliente? | ¿Va aparte? | Qué hicimos |
| --- | --- | --- | --- | --- |
| `datos/almacen_gorm.go:36` `Where(campo.nombre+" = ?", campo.valor)` | Sí | El valor sí (`?deporte=`, `?nivel=`). El nombre de la columna no: sale de una lista fija del código (`"deporte"`, `"nivel"`) | Sí, el valor con `?`. Además `negocio.ListarEventos` solo acepta valores de una lista (si no, 422) | Nada: es segura. El `+` pega un nombre de columna escrito en el código, no texto del cliente |
| `datos/almacen_gorm.go:40` `Where("sede_id IN (?)", … Where("LOWER(ciudad) = LOWER(?)", f.Ciudad))` | Sí | Sí, `?ciudad=` (texto libre, el único filtro sin lista) | Sí, con `?` | Nada. Es la que se atacó en el paso 1.4: devolvió `[]` |
| `datos/almacen_gorm.go:43` `Order("id").Find(&v)` | Sí | No: el orden está fijo | -- | Nada |
| `datos/almacen_gorm.go:48`, `:53`, `:58`, `:63` `First(&v, id)` | Sí | Sí, el `{id}` de la ruta (o un id ya guardado en la base) | `id` es `uint`: `idRuta` lo convierte con `strconv.ParseUint` y responde 400 si no es un entero positivo | Nada. `/inscripciones/1 OR 1=1` respondió 400 |
| `datos/almacen_gorm.go:89`, `:94` `Clauses(Locking…).First(&v, id)` | Sí | El `{id}` de la ruta | `id` es `uint` | Nada |
| `datos/almacen_gorm.go:68`, `:73`, `:78`, `:83`, `:99`, `:107` `Where("evento_id = ? …", …)` | Sí | Ids (`uint`) de la ruta o del token; los estados (`"confirmada"`, `"cancelada"`) son constantes del código | Sí, con `?` | Nada |
| `sesion/usuarios.go:86` `Where("nombre = ?", pedido.Usuario)` | Sí | Sí, el campo `usuario` del cuerpo JSON de `POST /login` | Sí, con `?` | Nada |
| `sesion/usuarios.go:55`, `datos/base.go:29`, `:38` `Where("nombre = ?", …)` | Sí | No: son los nombres fijos de la siembra (`ana`, `luis`) | Sí, con `?` | Nada |
| `datos/base.go:16` `Exec("CREATE UNIQUE INDEX IF NOT EXISTS …")` | Sí | No: texto fijo de la migración | -- | Nada |
| `datos/base.go:55` `Exec("SELECT setval(pg_get_serial_sequence('" + tabla + "', 'id'), …)")` | Sí, y pega texto con `+` | No: `tabla` sale de una lista fija del código (`usuarios`, `sedes`, `eventos`) y solo corre en la siembra (`DEMO_SEED=true`) | No, va pegado; un nombre de tabla no se puede pasar con `?` | Nada: no entra texto del cliente. Si algún día la lista viniera de afuera, habría que compararla contra una lista permitida |
| `datos/base.go:32` `nombre + "@eventosports.example"` | No: arma el correo de la siembra, que GORM envía como parámetro | -- | -- | Nada |
| `eventosports/manejadores.go:74` `q := r.URL.Query()` | No: lee la URL. De aquí salen `deporte`, `nivel` y `ciudad` hacia las filas de `almacen_gorm.go:36` y `:40` | -- | -- | Nada |
| `eventosports/rutas.go:24` `r.Group(…)`, `negocio/estados.go:37`, `sesion/guardia.go:60` y `main.go:44` | No: un grupo de rutas de chi, dos mensajes de error y la dirección del servidor | -- | -- | Nada |

**Resultado:** de las 27 líneas, 21 arman SQL y 6 no. El texto del cliente llega al SQL por cuatro
lugares: `?deporte=`, `?nivel=`, `?ciudad=` y el `usuario` del login, y en los cuatro va como parámetro
(`?`). Los `{id}` se convierten a `uint` antes de la consulta. El único `+` dentro de un SQL
(`base.go:55`) pega nombres de tabla fijos del código. Los cuerpos JSON de crear evento e inscripción
se guardan con `Create`, que también usa parámetros. No había nada que cerrar, así que no se cambió
código ni hace falta `auditoria_despues.txt`.

La prueba del paso 1.4 está en `inyeccion.txt`: `?ciudad=Manta` trae los eventos de Manta,
`?ciudad=' OR 1=1 --` devuelve `[]`, `?deporte=' OR 1=1 --` responde 422 y
`/inscripciones/1%20OR%201=1` responde 400. Lo mismo queda como prueba automática en
`servidor/internal/eventosports/inyeccion_test.go` (paso 1.4 opcional), que pasa con `go test ./...`.
