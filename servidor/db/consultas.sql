-- name: BuscarInscripcion :one
SELECT id, evento_id, jugador_id, estado, creado
FROM inscripciones
WHERE id = $1;

-- name: GuardarEstadoInscripcion :execrows
UPDATE inscripciones SET estado = $2 WHERE id = $1;
