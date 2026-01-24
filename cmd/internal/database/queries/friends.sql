-- name: CreateFriendRequest :one
INSERT INTO friendships (user_id, friend_id, status)
VALUES ($1, $2, 'pending')
RETURNING *;

-- name: AcceptFriendRequest :exec
UPDATE friendships
SET status = 'accepted', updated_at = NOW()
WHERE user_id = $1 AND friend_id = $2 AND status = 'pending';

-- name: RejectFriendRequest :exec
UPDATE friendships
SET status = 'rejected', updated_at = NOW()
WHERE user_id = $1 AND friend_id = $2 AND status = 'pending';

-- name: BlockUser :exec
INSERT INTO friendships (user_id, friend_id, status)
VALUES ($1, $2, 'blocked')
ON CONFLICT (user_id, friend_id)
DO UPDATE SET status = 'blocked', updated_at = NOW();

-- name: RemoveFriendship :exec
DELETE FROM friendships
WHERE (user_id = $1 AND friend_id = $2) OR (user_id = $2 AND friend_id = $1);

-- name: GetFriends :many
SELECT 
    CASE 
        WHEN f.user_id = $1 THEN f.friend_id
        ELSE f.user_id
    END as friend_id,
    u.email,
    u.created_at,
    f.created_at as friends_since
FROM friendships f
INNER JOIN users u ON (
    CASE 
        WHEN f.user_id = $1 THEN f.friend_id = u.id
        ELSE f.user_id = u.id
    END
)
WHERE (f.user_id = $1 OR f.friend_id = $1)
  AND f.status = 'accepted'
ORDER BY f.created_at DESC;

-- name: GetPendingRequests :many
SELECT u.id, u.email, u.created_at, f.created_at as requested_at
FROM friendships f
INNER JOIN users u ON f.user_id = u.id
WHERE f.friend_id = $1 AND f.status = 'pending'
ORDER BY f.created_at DESC;

-- name: GetSentRequests :many
SELECT u.id, u.email, u.created_at, f.created_at as requested_at
FROM friendships f
INNER JOIN users u ON f.friend_id = u.id
WHERE f.user_id = $1 AND f.status = 'pending'
ORDER BY f.created_at DESC;

-- name: AreFriends :one
SELECT EXISTS(
    SELECT 1 FROM friendships
    WHERE ((user_id = $1 AND friend_id = $2) OR (user_id = $2 AND friend_id = $1))
      AND status = 'accepted'
) as are_friends;

-- name: GetFriendshipStatus :one
SELECT status FROM friendships
WHERE (user_id = $1 AND friend_id = $2) OR (user_id = $2 AND friend_id = $1)
LIMIT 1;

-- name: GetFriendFeed :many
SELECT c.*, u.email as author_email
FROM chirps c
INNER JOIN users u ON c.user_id = u.id
WHERE c.user_id IN (
    SELECT 
        CASE 
            WHEN f.user_id = $1 THEN f.friend_id
            ELSE f.user_id
        END
    FROM friendships f
    WHERE (f.user_id = $1 OR f.friend_id = $1)
      AND f.status = 'accepted'
)
ORDER BY c.created_at DESC
LIMIT $2 OFFSET $3;