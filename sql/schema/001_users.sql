-- +goose Up
CREATE TABLE users (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    email TEXT NOT NULL UNIQUE,
    hashed_password TEXT NOT NULL DEFAULT 'unset',
    is_chirpy_red BOOLEAN DEFAULT FALSE
);

-- +goose Down
DROP TABLE users;

--postgres://postgres:postgres@localhost:5432/chirpy
--goose postgres postgres://postgres:postgres@localhost:5432/chirpy up
--psql -h localhost -U postgres -d chirpy