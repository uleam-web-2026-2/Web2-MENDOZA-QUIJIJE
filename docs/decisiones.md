# Decisiones de diseño — servidor de la mesa de ayuda (Semana 2)

Cada decisión sigue el formato: qué decidimos, por qué, qué alternativa descartamos y cómo lo comprobamos.

## D1 · Orden de los middleware

- **Decisión:** Registro afuera, Recuperación adentro (`r.Use(middleware.Registro); r.Use(middleware.Recuperacion)`). La cadena queda `Registro(Recuperacion(handler))`.
- **Por qué:** con `r.Use(A); r.Use(B)` el primero registrado es el más externo. Si un handler entra en pánico, `Recuperacion` (que está adentro) lo atrapa con `recover()`, responde 500 y termina normalmente. Como termina normalmente, el control vuelve a `Registro`, que sigue vivo por fuera, y este alcanza a escribir la línea `GET /explotar → 500` en el log. Así **toda** petición queda registrada, incluidas las que fallaron; para operar el servidor eso es lo más importante, porque un 500 que no aparece en el log es un fallo invisible.
- **Alternativa descartada:** el orden inverso (`Recuperacion` afuera, `Registro` adentro). El pánico sube desde el handler y atraviesa `Registro`, que no tiene `recover`, así que el código que va después de `next.ServeHTTP` (el `log.Printf`) nunca se ejecuta. `Recuperacion` lo atrapa por fuera y el cliente sí recibe el 500, pero esa petición no deja rastro en el log: el servidor respondió un error y el registro no se enteró.
- **Cómo lo comprobamos:** `curl -i localhost:8080/explotar` mirando la terminal del servidor. Con nuestro orden aparecen dos líneas: `PÁNICO en GET /explotar: runtime error: invalid memory address or nil pointer dereference` y a continuación `GET /explotar → 500 (0s)`. Con el orden inverso solo aparece la línea de `PÁNICO`; la de `→ 500` no se escribe.

## D2 · Código para JSON roto frente a datos inválidos

- **Decisión:** códigos distintos. **400 Bad Request** con `cuerpo_malformado` cuando el cuerpo no se puede leer como JSON; **422 Unprocessable Entity** con `datos_invalidos` cuando el JSON es válido pero viola una regla de negocio (título de menos de 5 caracteres, prioridad fuera de baja/media/alta).
- **Por qué:** son dos fallos de naturaleza distinta y el cliente los arregla en lugares distintos. Un 400 le dice «no te entendí, arregla *cómo* envías» (el problema está en la serialización, en el código que arma la petición). Un 422 le dice «te entendí perfectamente, arregla *qué* envías» (el problema está en los datos que el usuario escribió: le falta texto al título, eligió una prioridad que no existe). Con un solo código el cliente tendría que leer el `mensaje` para distinguirlos, y el mensaje es texto para personas, no para programas. Con dos códigos puede, por ejemplo, mostrar el 422 como error de formulario al usuario y tratar el 400 como bug propio.
- **Alternativa descartada:** 400 para ambos. Es válido y se usa en la industria, pero mezcla «petición mal construida» con «datos incorrectos» y obliga al cliente a inspeccionar el mensaje. Preferimos que el código HTTP ya cuente la historia.
- **Cómo lo comprobamos:** `curl -i -X POST localhost:8080/tickets -d '{"titulo": "x'` → `400` `cuerpo_malformado`. `curl -i -X POST localhost:8080/tickets -d '{"titulo":"abc","prioridad":"alta"}'` → `422` `datos_invalidos`. `curl -i -X POST localhost:8080/tickets -d '{"titulo":"Teclado dañado","prioridad":"urgente"}'` → `422` `datos_invalidos`. Los tres casos están en `pruebas.sh`.

## D3 · Campo desconocido en el cuerpo

- **Decisión:** rechazar el cuerpo con **400** `cuerpo_malformado`. Se activa `dec.DisallowUnknownFields()` antes de `Decode`, así un JSON con un campo que no sea `titulo` o `prioridad` hace fallar la decodificación.
- **Por qué:** el caso típico es `{"titulo":"Pantalla rota","prioridad":"alta","estado":"cerrado"}`: el cliente intenta fijar el `estado`, que solo el servidor decide (todo ticket nace `abierto`). Si lo ignoráramos en silencio, el ticket se crearía abierto y el cliente creería que lo creó cerrado; el error quedaría oculto hasta que alguien lo notara. Rechazarlo hace visible el intento en el momento y obliga a corregir el cliente. Además protege contra errores de tipeo (`"prioridd"`): en vez de crear un ticket con prioridad vacía que después falla en la validación con un mensaje confuso, el servidor dice claramente que el cuerpo no es el esperado.
- **Alternativa descartada:** ignorar el campo (comportamiento por defecto de `encoding/json`). Es más tolerante con clientes viejos que mandan campos extra, pero en este sistema no hay clientes viejos y preferimos que el contrato sea estricto desde el inicio.
- **Cómo lo comprobamos:** `curl -i -X POST localhost:8080/tickets -d '{"titulo":"Pantalla rota","prioridad":"alta","estado":"cerrado"}'` → `400` `cuerpo_malformado`. Sin `DisallowUnknownFields` respondía `201` con `"estado":"abierto"`.

## D4 · Límite de tamaño del cuerpo

- **Decisión:** tope de **1 MB** por cuerpo con `io.LimitReader(r.Body, 1<<20)`. Un cuerpo que lo excede entra por el mismo camino que el JSON roto: **400** `cuerpo_malformado`.
- **Por qué:** sin límite, `json.NewDecoder(r.Body)` lee lo que el cliente mande. Un título de 2 GB se guardaría en memoria; con pocos clientes así el servidor se queda sin RAM y se cae para todos. 1 MB es cientos de veces más de lo que un título razonable necesita y aun así protege al servidor.
- **Alternativa descartada:** `http.MaxBytesReader` con un **413 Content Too Large** propio. Sería más preciso para el cliente, pero agrega un código más al contrato y una rama más de detección de error. Para este taller preferimos el contrato más simple: «si no pude leer el cuerpo, es 400». Se puede cambiar a 413 más adelante sin tocar nada más.
- **Cómo lo comprobamos:** `pruebas.sh` arma un JSON con un título de 2 MB y lo envía con `--data-binary`. Responde `400`; sin el límite respondía `201`. Después, `GET /tickets` responde `200`: el servidor sigue vivo.

## D5 · El título se recorta antes de validar

- **Decisión:** `strings.TrimSpace(entrada.Titulo)` antes de medir la longitud y antes de guardar.
- **Por qué:** un título de `"     ab   "` tiene 10 caracteres pero solo 2 con contenido. Sin recorte pasaría la regla de mínimo 5 y se guardaría un ticket con espacios basura. Además el ticket se guarda ya limpio, así lo que se lista es lo que el usuario realmente escribió.
- **Alternativa descartada:** validar el texto tal cual llega. Más simple, pero deja pasar títulos que en la práctica están vacíos.
- **Cómo lo comprobamos:** `curl -i -X POST localhost:8080/tickets -d '{"titulo":"    ab    ","prioridad":"alta"}'` → `422` `datos_invalidos`.
