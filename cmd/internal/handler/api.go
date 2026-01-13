package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/shubh-man007/Chirpy/cmd/internal/database"
)

var profane = []string{"kerfuffle", "sharbert", "fornax"}
var astrix = "****"

type APIHandler struct {
	db *database.Queries
}

func NewAPIHandler(db *database.Queries) *APIHandler {
	return &APIHandler{
		db: db,
	}
}

type ChirpBody struct {
	Body string `json:"body"`
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

func ValidateChirp(w http.ResponseWriter, r *http.Request) {
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

	chirpData := chirp.Body

	for _, prof := range profane {
		if strings.Contains(chirpData, prof) {
			chirpData = strings.ReplaceAll(chirpData, prof, astrix)
		}
	}

	if utf8.RuneCountInString(chirpData) > 140 {
		errJSON(w, http.StatusBadRequest, ErrMessage{
			Message: "Chirp too long",
		})
		return
	}

	respondJSON(w, http.StatusOK, ChirpLenValid{
		Body:    chirpData,
		Message: true,
	})
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

	user, err := h.db.CreateUser(r.Context(), req.Email)
	if err != nil {
		log.Printf("Error creating user: %s", err)
		errJSON(w, http.StatusInternalServerError, ErrMessage{
			Message: "Could not create user",
		})
		return
	}

	respondJSON(w, http.StatusCreated, user)
}
