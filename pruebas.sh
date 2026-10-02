#!/usr/bin/env bash
# pruebas.sh — recorre todo el contrato de la API con curl.
# Cada caso imprime el código HTTP esperado y, debajo, lo que respondió el servidor,
# para comparar a ojo. El servidor debe estar corriendo antes (go run .).
#
# Uso: ./pruebas.sh                              (contra localhost:8080)
#      ./pruebas.sh http://192.168.1.20:8080      (contra otra pareja)

# B = URL base. ${1:-...} usa el primer argumento si existe; si no, localhost:8080.
B=${1:-http://localhost:8080}

# c = "caso". Recibe: nombre del caso, código esperado, y el resto son argumentos para curl.
#   curl -s -i  → silencioso, pero incluye las cabeceras de respuesta
#   sed -n '1p;/^{/p' → de toda la salida muestra solo la línea 1 (HTTP/1.1 200 OK)
#                        y las líneas que empiezan con "{" (el JSON del cuerpo)
c() { echo; echo "### $1 [esperado: $2]"; shift 2
  curl -s -i "$@" | sed -n '1p;/^{/p'; }

# ---------- Casos felices ----------

c "POST válido" "201" -X POST "$B/tickets" \
  -d '{"titulo":"Impresora sin tóner","prioridad":"alta"}'

c "GET lista" "200" "$B/tickets"

c "GET por id" "200" "$B/tickets/1"

# ---------- Errores del cliente: 400 (no te entendí) vs 422 (te entendí, pero no es válido) ----------
c "GET id inexistente" "404" "$B/tickets/99"

c "GET id no numérico" "400" "$B/tickets/abc"

c "POST JSON roto" "400" -X POST "$B/tickets" -d '{"titulo": "x'

c "POST título corto" "422" -X POST "$B/tickets" \
  -d '{"titulo":"abc","prioridad":"alta"}'

c "POST prioridad mala" "422" -X POST "$B/tickets" \
  -d '{"titulo":"Teclado dañado","prioridad":"urgente"}'

# ---------- Errores del router: 405 y 404 con la misma envoltura JSON ----------
c "DELETE no soportado" "405" -X DELETE "$B/tickets/1"

c "ruta inexistente" "404" "$B/usuarios"

# ---------- Middleware Recuperacion: el pánico se convierte en 500 y el servidor no muere ----------
c "pánico controlado" "500" "$B/explotar"

# ---------- Protecciones de la Fase 3: campos desconocidos y cuerpo gigante ----------
c "campo desconocido" "400" -X POST "$B/tickets" \
  -d '{"titulo":"Pantalla rota","prioridad":"alta","estado":"cerrado"}'

# Cuerpo gigante (2 MB): se arma un archivo aparte porque no cabe en la línea de comandos.
printf '{"titulo":"' > /tmp/grande.json
head -c 2000000 /dev/zero | tr '\0' 'A' >> /tmp/grande.json
printf '","prioridad":"alta"}' >> /tmp/grande.json
echo
echo "### cuerpo de 2 MB [esperado: 400]"
curl -s -o /dev/null -w "%{http_code}\n" -X POST "$B/tickets" --data-binary @/tmp/grande.json

# Si esto responde 200, el pánico y el cuerpo gigante no tumbaron el servidor.
c "el servidor sigue vivo" "200" "$B/tickets"