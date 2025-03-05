-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, user_id, expires_at, revoked_at)
VALUES ( 
    $1,                 -- Body передается как параметр
    NOW(),              -- Текущее время
    NOW(),              -- Текущее время
    $2,                 -- User_id передается как параметр
    $3,
    NULL
)
RETURNING *;

-- name: DeleteRefreshToken :exec
DELETE FROM refresh_tokens;

-- name: RevokeToken :exec
UPDATE refresh_tokens
SET updated_at = NOW(), revoked_at = NOW()
WHERE token = $1;

-- name: GetUserFromRefreshToken :one
SELECT * FROM refresh_tokens
WHERE token = $1 LIMIT 1;
