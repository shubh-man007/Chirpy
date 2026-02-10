-- name: FollowUser :exec
INSERT INTO follows (follower_id, followee_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: UnfollowUser :exec
DELETE FROM follows
WHERE follower_id = $1 AND followee_id = $2;

-- name: GetFollowers :many
SELECT u.id, u.email, u.is_chirpy_red, u.created_at, f.created_at as followed_at
FROM follows f
INNER JOIN users u ON f.follower_id = u.id
WHERE f.followee_id = $1
ORDER BY f.created_at DESC;

-- name: GetFollowing :many
SELECT u.id, u.email, u.is_chirpy_red, u.created_at, f.created_at as followed_at
FROM follows f
INNER JOIN users u ON f.followee_id = u.id
WHERE f.follower_id = $1
ORDER BY f.created_at DESC;

-- name: IsFollowing :one
SELECT EXISTS(
    SELECT 1 FROM follows
    WHERE follower_id = $1 AND followee_id = $2
) as is_following;

-- name: GetFollowerCount :one
SELECT COUNT(*) FROM follows WHERE followee_id = $1;

-- name: GetFollowingCount :one
SELECT COUNT(*) FROM follows WHERE follower_id = $1;

-- name: GetFeed :many
SELECT c.id, c.created_at, c.updated_at, c.body, c.user_id, u.email as author_email
FROM chirps c
INNER JOIN users u ON c.user_id = u.id
WHERE EXISTS (
    SELECT 1
    FROM follows f
    WHERE f.follower_id = $1
      AND f.followee_id = c.user_id
)
ORDER BY c.created_at DESC
LIMIT $2 OFFSET $3;