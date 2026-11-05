CREATE TABLE inscripciones (
    id bigserial PRIMARY KEY,
    evento_id bigint NOT NULL,
    jugador_id bigint NOT NULL,
    estado text NOT NULL DEFAULT 'reservada',
    creado timestamptz
);

CREATE TABLE pagos (
    id bigserial PRIMARY KEY,
    inscripcion_id bigint NOT NULL REFERENCES inscripciones(id),
    monto numeric(8,2) NOT NULL,
    banco text NOT NULL,
    referencia text NOT NULL,
    verificado boolean NOT NULL DEFAULT false,
    creado timestamptz NOT NULL
);
