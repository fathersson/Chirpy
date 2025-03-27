-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password, is_chirpy_red)
VALUES (
    gen_random_uuid(),  -- Генерация нового UUID
    NOW(),              -- Текущее время
    NOW(),              -- Текущее время
    $1,                  -- Email передается как параметр
    $2,
    FALSE
)
RETURNING *;

-- name: DeleteUsers :exec
DELETE FROM users;

-- name: GetUser :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: UpdateEMailPasword :one
UPDATE users
SET email = $1, hashed_password = $2
WHERE id = $3
RETURNING *;

-- name: UpdateChirpyRed :one
UPDATE users
SET  is_chirpy_red = TRUE
WHERE id = $1
RETURNING *;

-- name: CheckUser :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;