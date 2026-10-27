// Recorrido reproducible: crea datos nuevos en cada ejecucion.
// Solo usa usuarios y claves inventados del taller. Nunca imprime los JWT.
const base = process.argv[2] || "http://localhost:8080";
const reporte = [];
async function enviar(metodo, ruta, codigo, cuerpo, token, etiqueta) {
  const response = await fetch(base + ruta, {
    method: metodo,
    headers: { ...(cuerpo ? { "Content-Type": "application/json" } : {}), ...(token ? { Authorization: `Bearer ${token}` } : {}) },
    body: cuerpo ? JSON.stringify(cuerpo) : undefined,
  });
  const data = await response.json();
  if (response.status !== codigo) throw new Error(`${metodo} ${ruta}: esperaba ${codigo}, recibio ${response.status}: ${JSON.stringify(data)}`);
  if (etiqueta) { const fila = { paso: etiqueta, metodo, ruta, codigo: response.status }; reporte.push(fila); console.log(`${etiqueta}: ${metodo} ${ruta} -> ${response.status}`); }
  return data;
}
async function login(usuario, clave, etiqueta) { return (await enviar("POST", "/login", 200, { usuario, clave }, null, etiqueta)).token; }
const jugador = await login("ana", "ana-2026");
const organizadorPrevio = await login("luis", "luis-2026");
const nuevoEvento = (precio) => ({ titulo: `Ensayo S7 ${Date.now()}`, deporte: "futbol", nivel: "casual", inicio: new Date(Date.now() + 7 * 86400000).toISOString(), cupos: 14, precio, sede_id: 3 });
// Preparacion: una reserva pendiente de verificar y otra cancelada.
const evento = await enviar("POST", "/eventos", 201, nuevoEvento(2.5), organizadorPrevio);
const reserva = await enviar("POST", `/eventos/${evento.id}/inscripciones`, 201, null, jugador);
const ruta = `/inscripciones/${reserva.id}/estado`;
await enviar("PATCH", ruta, 200, { estado: "pago_por_verificar", banco: "Banco de ejemplo", referencia: `S7-${Date.now()}` }, jugador);
const eventoFinal = await enviar("POST", "/eventos", 201, nuevoEvento(0), organizadorPrevio);
const cancelada = await enviar("POST", `/eventos/${eventoFinal.id}/inscripciones`, 201, null, jugador);
const rutaFinal = `/inscripciones/${cancelada.id}/estado`;
await enviar("PATCH", rutaFinal, 200, { estado: "cancelada" }, jugador);
console.log("Preparacion: dos eventos y dos inscripciones nuevos; no se borraron datos.");
await enviar("GET", `/eventos/${evento.id}/inscripciones`, 401, null, null, "1. Sin token");
const organizador = await login("luis", "luis-2026", "2. Login");
await enviar("GET", `/eventos/${evento.id}/inscripciones`, 200, null, organizador, "3. Con token");
await enviar("PATCH", ruta, 403, { estado: "confirmada" }, jugador, "4. Rol sin permiso");
await enviar("PATCH", ruta, 200, { estado: "confirmada" }, organizador, "5. Rol permitido");
await enviar("PATCH", rutaFinal, 422, { estado: "confirmada" }, organizador, "6. Transicion prohibida");
if (process.argv[3]) { const { writeFile } = await import("node:fs/promises"); await writeFile(process.argv[3], JSON.stringify({ fecha: new Date().toISOString(), base, resultado: "OK", peticiones: reporte, eventoId: evento.id, inscripcionId: reserva.id }, null, 2) + "\n"); }
console.log("Recorrido S7 correcto. Los tokens no se guardan en el reporte.");
