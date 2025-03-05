-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
VALUES ( 
    gen_random_uuid(),  -- Генерация нового UUID
    NOW(),              -- Текущее время
    NOW(),              -- Текущее время
    $1,                 -- Body передается как параметр
    $2                  -- User_id передается как параметр

)
RETURNING *;

-- name: DeleteChirps :exec
DELETE FROM chirps;

-- name: DeleteChirp :exec
DELETE FROM chirps
WHERE id = $1;

-- name: GetChirps :many
SELECT * FROM chirps
ORDER BY created_at ASC;

-- name: GetChirpsAuthorID :many
SELECT * FROM chirps
WHERE user_id = $1
ORDER BY created_at ASC;


-- name: GetChirp :one
SELECT * FROM chirps
WHERE id = $1 LIMIT 1;

