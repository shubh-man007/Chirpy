package models

import "time"

type ProfileResponse struct {
	ID             string      `json:"id"`
	Email          string      `json:"email"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
	IsChirpyRed    bool        `json:"is_chirpy_red"`
	FollowersCount int64       `json:"followers_count"`
	FollowingCount int64       `json:"following_count"`
	ChirpsCount    int64       `json:"chirps_count"`
	Chirps         []ChirpItem `json:"chirps"`
	NextCursor     *string     `json:"next_cursor,omitempty"`
}

type ChirpItem struct {
	ID        string    `json:"id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}
