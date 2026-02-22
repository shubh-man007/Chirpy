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

-- name: DeleteUser :exec
DELETE FROM users;

-- name: DeleteUserByID :exec
DELETE FROM users WHERE id = $1;

-- name: GetUserProfile :one
WITH user_stats AS (
    SELECT 
        u.id,
        u.email,
        u.created_at,
        u.updated_at,
        u.is_chirpy_red,
        (SELECT COUNT(*) FROM follows WHERE followee_id = u.id) as followers_count,
        (SELECT COUNT(*) FROM follows WHERE follower_id = u.id) as following_count,
        (SELECT COUNT(*) FROM chirps WHERE user_id = u.id) as chirps_count
    FROM users u
    WHERE u.id = $1
)
SELECT * FROM user_stats;

-- name: GetUserChirpsPaginated :many
SELECT 
    c.id,
    c.created_at,
    c.updated_at,
    c.body,
    c.user_id
FROM chirps c
WHERE c.user_id = sqlc.arg(user_id)
AND (
    sqlc.narg(cursor)::uuid IS NULL OR 
    (c.created_at, c.id) < (
        SELECT created_at, id FROM chirps WHERE id = sqlc.narg(cursor)
    )
)
ORDER BY c.created_at DESC, c.id DESC
LIMIT sqlc.arg(page_limit);