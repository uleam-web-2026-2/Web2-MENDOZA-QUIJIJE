-- name: BuscarInscripcion :one
SELECT id, evento_id, jugador_id, estado, creado
FROM inscripciones
WHERE id = $1;

-- name: GuardarEstadoInscripcion :execrows
UPDATE inscripciones SET estado = $2 WHERE id = $1;

-- name: CrearPago :one
INSERT INTO pagos (inscripcion_id, monto, banco, referencia, verificado, creado)
VALUES (sqlc.arg(inscripcion_id), sqlc.arg(monto)::double precision,
        sqlc.arg(banco), sqlc.arg(referencia), sqlc.arg(verificado), sqlc.arg(creado))
RETURNING id;
