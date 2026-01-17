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
	"github.com/shubh-man007/Chirpy/cmd/internal/auth"
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

type UserLogin struct {
	Password string `json:"password"`
	Email    string `json:"email"`
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
	var req UserLogin
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request JSON: %s", err)
		errJSON(w, http.StatusBadRequest, ErrMessage{
			Message: "Invalid request body",
		})
		return
	}

	hashed_password, err := auth.HashPassword(req.Password)
	if err != nil {
		log.Printf("Error hashing password: %s", err)
		errJSON(w, http.StatusInternalServerError, ErrMessage{
			Message: "Error hashing",
		})
		return
	}

	userParams := database.CreateUserParams{
		Email:          req.Email,
		HashedPassword: hashed_password,
	}
	user, err := h.cfg.DB.CreateUser(r.Context(), userParams)
	if err != nil {
		log.Printf("Error creating user: %s", err)
		http.Error(w, fmt.Sprintf("error creating user: %v", err), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusCreated, user)
}

func (h *APIHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var req UserLogin
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request JSON: %s", err)
		errJSON(w, http.StatusBadRequest, ErrMessage{
			Message: "Invalid request body",
		})
		return
	}

	userCreds, err := h.cfg.DB.GetUserPassByEmail(r.Context(), req.Email)
	if err != nil {
		log.Printf("Error validating user: %s", err)
		http.Error(w, fmt.Sprintf("Error validating user: %v", err), http.StatusInternalServerError)
		return
	}

	val, err := auth.CheckPasswordHash(req.Password, userCreds)
	if err != nil {
		log.Printf("Unauthorized User: %s", err)
		http.Error(w, "Unauthorized user", http.StatusUnauthorized)
		return
	}

	if !val {
		log.Printf("Unauthorized User: %s", err)
		http.Error(w, "Unauthorized user", http.StatusUnauthorized)
		return
	}

	user, err := h.cfg.DB.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		log.Printf("Error fetching user: %s", err)
		http.Error(w, fmt.Sprintf("error fetching chirps: %v", err), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, user)
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

func (h *APIHandler) GetChirpsByUser(w http.ResponseWriter, r *http.Request) {
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		log.Printf("Could not parse chirp ID: %s", err)
		errJSON(w, http.StatusBadRequest, ErrMessage{
			Message: "Could not parse chirp ID",
		})
		return
	}
	chirps, err := h.cfg.DB.GetChirpsByUser(r.Context(), chirpID)
	if err != nil {
		log.Printf("Error fetching chirps: %s", err)
		http.Error(w, fmt.Sprintf("error fetching chirps: %v", err), http.StatusInternalServerError)
		return
	}

	if len(chirps) == 0 {
		errJSON(w, http.StatusNotFound, ErrMessage{
			Message: "No chirps found for this user",
		})
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
