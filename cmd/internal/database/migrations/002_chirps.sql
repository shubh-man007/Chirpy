-- +goose Up
CREATE TABLE chirps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    body TEXT NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_chirps_user_id ON chirps(user_id);
CREATE INDEX idx_chirps_created_at ON chirps(created_at DESC);

-- +goose Down
DROP TABLE chirps;

-- up migration: goose -dir cmd/internal/database/migrations postgres "postgres://postgres:chirpDB@localhost:5431/chirpy?sslmode=disable" up