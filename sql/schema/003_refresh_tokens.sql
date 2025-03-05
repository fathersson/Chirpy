-- +goose Up
CREATE TABLE refresh_tokens (
    token VARCHAR(64) PRIMARY KEY,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP DEFAULT NULL --метка времени, когда токен был отозван (null, если не отозван)
);

-- +goose Down
DROP TABLE refresh_tokens;