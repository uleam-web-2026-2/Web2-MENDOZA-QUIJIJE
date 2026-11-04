-- Tabla deportes
CREATE TABLE IF NOT EXISTS deportes (
    id SERIAL PRIMARY KEY,
    nombre VARCHAR(100) NOT NULL UNIQUE
);

-- Tabla eventos
CREATE TABLE IF NOT EXISTS eventos (
    id SERIAL PRIMARY KEY,
    titulo VARCHAR(150) NOT NULL,
    descripcion TEXT,
    deporte_id INT REFERENCES deportes(id),
    fecha TIMESTAMP NOT NULL
);