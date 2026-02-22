package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/shubh-man007/Chirpy/cmd/internal/config"
)

type AdminHandler struct {
	apiCfg *config.ApiConfig
}

func NewAdminHandler(cfg *config.ApiConfig) *AdminHandler {
	return &AdminHandler{apiCfg: cfg}
}

type APIHandler struct {
	cfg *config.ApiConfig
}

func NewAPIHandler(cfg *config.ApiConfig) *APIHandler {
	return &APIHandler{
		cfg: cfg,
	}
}

type UserLogin struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

type ProfileResponse struct {
	ID             uuid.UUID   `json:"id"`
	Email          string      `json:"email"`
	CreatedAt      string      `json:"created_at"`
	UpdatedAt      string      `json:"updated_at"`
	IsChirpyRed    bool        `json:"is_chirpy_red"`
	FollowersCount int64       `json:"followers_count"`
	FollowingCount int64       `json:"following_count"`
	ChirpsCount    int64       `json:"chirps_count"`
	Chirps         []ChirpItem `json:"chirps"`
	NextCursor     *string     `json:"next_cursor,omitempty"`
	IsFollowing    *bool       `json:"is_following,omitempty"`
}

type ChirpItem struct {
	ID        uuid.UUID `json:"id"`
	Body      string    `json:"body"`
	CreatedAt string    `json:"created_at"`
}

type ChirpBody struct {
	Body string `json:"body"`
}

type ChirpLenValid struct {
	Body    string `json:"cleaned_body"`
	Message bool   `json:"valid"`
}

type MembershipWebhookData struct {
	UserID string `json:"user_id"`
}

type MembershipWebhookEvent struct {
	Event string                `json:"event"`
	Data  MembershipWebhookData `json:"data"`
}

type ErrMessage struct {
	Message string `json:"error"`
}

func errJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("Error encoding response: %s", err)
	}
}

func respondJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("Error encoding response: %s", err)
	}
}
