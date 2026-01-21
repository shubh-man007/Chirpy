-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING id, created_at, updated_at, email, is_chirpy_red;

-- name: GetUserByID :one
SELECT id, created_at, updated_at, email, is_chirpy_red FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT id, created_at, updated_at, email, is_chirpy_red FROM users WHERE email = $1;

-- name: GetUserPassByEmail :one
SELECT hashed_password FROM users WHERE email = $1;

-- name: DeleteUser :exec
DELETE FROM users;

-- name: UpdateUserCred :one
UPDATE users
SET updated_at = NOW(),
    email = $1,
    hashed_password = $2
WHERE id = $3
RETURNING id, created_at, updated_at, email, is_chirpy_red;

-- name: UpdateUserToChirpyRed :exec
UPDATE users
SET is_chirpy_red = TRUE,
    updated_at = NOW()
WHERE id = $1;