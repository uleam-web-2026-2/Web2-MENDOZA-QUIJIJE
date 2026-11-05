-- name: ListDeportes :many
SELECT * FROM deportes
ORDER BY id;
  
-- name: GetEvento :one
SELECT * FROM eventos
WHERE id = $1 LIMIT 1;

-- name: ListEventos :many
SELECT * FROM eventos
ORDER BY fecha DESC;

-- name: CreateEvento :one
INSERT INTO eventos (
  titulo, descripcion, deporte_id, fecha
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: UpdateEvento :one
UPDATE eventos
SET titulo = $1,
    descripcion = $2,
    deporte_id = $3,
    fecha = $4
WHERE id = $5
RETURNING *;

-- name: DeleteEvento :exec
DELETE FROM eventos
WHERE id = $1;