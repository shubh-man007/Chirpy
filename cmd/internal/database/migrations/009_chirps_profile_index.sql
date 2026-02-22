-- +goose Up
CREATE INDEX IF NOT EXISTS idx_chirps_user_created_id ON chirps(user_id, created_at DESC, id DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_chirps_user_created_id;
