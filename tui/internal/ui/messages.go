package ui

import (
	"github.com/shubh-man007/Chirpy/tui/internal/models"
)

type LoginSuccessMsg struct {
	User *models.User
}

type LoginFailureMsg struct {
	Err error
}

type ErrorMsg struct {
	Err error
}

type FeedLoadedMsg struct {
	Chirps []models.Chirp
	Append bool
}

type ChirpPostedMsg struct {
	Chirp *models.Chirp
}

type UserUpdatedMsg struct {
	User *models.User
	Err  error
}

type UserDeletedMsg struct {
	Err error
}

type UserChirpsLoadedMsg struct {
	Chirps []models.Chirp
	Err    error
}

type ProfileLoadedMsg struct {
	Profile *models.ProfileResponse
	Err     error
}

type FollowersLoadedMsg struct {
	Followers []models.FollowerRow
	Err       error
}

type FollowingLoadedMsg struct {
	Following []models.FollowingRow
	Err       error
}

type FollowUnfollowSuccessMsg struct {
	UserID string
}
