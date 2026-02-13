package models

import "time"

type FeedChirp struct {
	Chirps []Chirp `json:"feed,omitempty"`
}

type FollowResponse struct {
	Followers []FollowUser `json:"followers,omitempty"`
	Following []FollowUser `json:"following,omitempty"`
	Total     int64        `json:"total"`
}

type FollowUser struct {
	ID          string    `json:"id"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
	FollowedAt  time.Time `json:"followed_at"`
}
