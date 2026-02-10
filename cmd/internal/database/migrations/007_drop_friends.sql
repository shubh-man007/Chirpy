-- +goose Up
DROP TABLE IF EXISTS friendships;

-- +goose Down
CREATE TABLE friendships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    friend_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL CHECK (status IN ('pending', 'accepted', 'rejected', 'blocked')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Prevent duplicate friend requests
    UNIQUE(user_id, friend_id),
    
    -- Prevent self-friending
    CHECK (user_id != friend_id)
);

-- Indexes for fast lookups
CREATE INDEX idx_friendships_user_id ON friendships(user_id);
CREATE INDEX idx_friendships_friend_id ON friendships(friend_id);
CREATE INDEX idx_friendships_status ON friendships(status);

-- Composite index for common queries
CREATE INDEX idx_friendships_user_status ON friendships(user_id, status);