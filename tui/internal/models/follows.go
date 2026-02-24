package models

import "time"

type Follow struct {
	FolloweeID string `json:"followee_id"`
}

type FollowResponse struct {
	Followers []FollowUser `json:"followers,omitempty"`
	Following []FollowUser `json:"following,omitempty"`
	Total     int64        `json:"total"`
}

type FollowersResponse struct {
	Followers     []FollowerRow `json:"Followers"`
	FollowerCount int64         `json:"FollowerCount"`
}

type FollowingResponse struct {
	Following      []FollowingRow `json:"Following"`
	FollowingCount int64          `json:"FollowingCount"`
}

type FollowerRow struct {
	ID             string    `json:"id"`
	Email          string    `json:"email"`
	IsChirpyRed    bool      `json:"is_chirpy_red"`
	CreatedAt      time.Time `json:"created_at"`
	FollowedAt     time.Time `json:"followed_at"`
	TotalFollowers int64     `json:"total_followers"`
}

type FollowingRow struct {
	ID             string    `json:"id"`
	Email          string    `json:"email"`
	IsChirpyRed    bool      `json:"is_chirpy_red"`
	CreatedAt      time.Time `json:"created_at"`
	FollowedAt     time.Time `json:"followed_at"`
	TotalFollowing int64     `json:"total_following"`
}

type FollowUser struct {
	ID          string    `json:"id"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
	FollowedAt  time.Time `json:"followed_at"`
}
