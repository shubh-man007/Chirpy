package models

type LoginResponse struct {
	UserID       string `json:"user_id"`
	IsChirpyRed  bool   `json:"is_chirpy_red"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}
