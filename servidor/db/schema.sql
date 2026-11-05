CREATE TABLE inscripciones (
    id bigserial PRIMARY KEY,
    evento_id bigint NOT NULL,
    jugador_id bigint NOT NULL,
    estado text NOT NULL DEFAULT 'reservada',
    creado timestamptz
);
