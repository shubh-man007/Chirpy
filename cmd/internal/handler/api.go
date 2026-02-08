package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/shubh-man007/Chirpy/cmd/internal/config"
)

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
