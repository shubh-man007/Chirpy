package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/shubh-man007/Chirpy/cmd/internal/config"
	"github.com/shubh-man007/Chirpy/cmd/internal/database"
)

const (
	maxChirpLength = 140
	asterisk       = "****"
)

var profane = []string{"kerfuffle", "sharbert", "fornax"}

type APIHandler struct {
	cfg *config.ApiConfig
}

func NewAPIHandler(cfg *config.ApiConfig) *APIHandler {
	return &APIHandler{
		cfg: cfg,
	}
}

type ChirpBody struct {
	Body   string `json:"body"`
	UserID string `json:"user_id"`
}

type ChirpLenValid struct {
	Body    string `json:"cleaned_body"`
	Message bool   `json:"valid"`
}

type ErrMessage struct {
	Message string `json:"error"`
}

type userEmail struct {
	Email string `json:"email"`
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

func (h *APIHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req userEmail
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request JSON: %s", err)
		errJSON(w, http.StatusBadRequest, ErrMessage{
			Message: "Invalid request body",
		})
		return
	}

	user, err := h.cfg.DB.CreateUser(r.Context(), req.Email)
	if err != nil {
		log.Printf("Error creating user: %s", err)
		http.Error(w, fmt.Sprintf("error creating user: %v", err), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusCreated, user)
}

func (h *APIHandler) CreateChirp(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	chirp := ChirpBody{}
	err := decoder.Decode(&chirp)
	if err != nil {
		log.Printf("Error decoding requested JSON: %s", err)
		errJSON(w, http.StatusInternalServerError, ErrMessage{
			Message: "Something went wrong",
		})
		return
	}

	if utf8.RuneCountInString(chirp.Body) > maxChirpLength {
		errJSON(w, http.StatusBadRequest, ErrMessage{
			Message: "Chirp too long",
		})
		return
	}

	cleanChirpBody := cleanProfanity(chirp.Body)

	userID, err := uuid.Parse(chirp.UserID)
	if err != nil {
		log.Printf("Could not parse User ID: %s", err)
		errJSON(w, http.StatusBadRequest, ErrMessage{
			Message: "Could not parse User ID",
		})
		return
	}

	valChirp, err := h.cfg.DB.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleanChirpBody,
		UserID: userID,
	})

	if err != nil {
		log.Printf("Error creating chirp: %s", err)
		http.Error(w, fmt.Sprintf("error creating chirp: %v", err), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusCreated, valChirp)
}

func (h *APIHandler) GetAllChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := h.cfg.DB.GetAllChirps(r.Context())
	if err != nil {
		log.Printf("Error fetching chirps: %s", err)
		http.Error(w, fmt.Sprintf("error fetching chirps: %v", err), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, chirps)
}

// utility:
func cleanProfanity(text string) string {
	words := strings.Split(text, " ")
	for i, word := range words {
		cleanWord := strings.ToLower(strings.Trim(word, ".,!?;:"))
		if slices.Contains(profane, cleanWord) {
			words[i] = asterisk
		}
	}
	return strings.Join(words, " ")
}
