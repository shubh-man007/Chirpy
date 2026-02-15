package models

type LoginCreds struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}
type LoginResponse struct {
	UserID       string `json:"user_id"`
	IsChirpyRed  bool   `json:"is_chirpy_red"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}
