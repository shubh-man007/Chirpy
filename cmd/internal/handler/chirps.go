package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/shubh-man007/Chirpy/cmd/internal/auth"
	"github.com/shubh-man007/Chirpy/cmd/internal/database"
)

const (
	maxChirpLength = 140
	asterisk       = "****"
)

var profane = []string{"kerfuffle", "sharbert", "fornax"}

func (h *APIHandler) CreateChirp(w http.ResponseWriter, r *http.Request) {
	chirp := ChirpBody{}
	err := json.NewDecoder(r.Body).Decode(&chirp)
	if err != nil {
		log.Printf("Error decoding requested JSON: %v", err)
		errJSON(w, http.StatusInternalServerError, ErrMessage{
			Message: "Something went wrong",
		})
		return
	}

	sentJWTToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Invalid Token: %v", err)
		errJSON(w, http.StatusUnauthorized, ErrMessage{
			Message: "Unauthorized",
		})
		return
	}

	userID, err := auth.ValidateJWT(sentJWTToken, h.cfg.JWTSecret)
	if err != nil {
		log.Printf("Invalid Token: %v", err)
		errJSON(w, http.StatusUnauthorized, ErrMessage{
			Message: "Invalid Token",
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

	valChirp, err := h.cfg.DB.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleanChirpBody,
		UserID: userID,
	})

	if err != nil {
		log.Printf("Error creating chirp: %v", err)
		http.Error(w, "Couldn't chirp", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusCreated, valChirp)
}

func (h *APIHandler) GetAllChirps(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	sortOrder := query.Get("sort")

	if sortOrder == "" {
		sortOrder = "asc"
	}

	chirps, err := h.cfg.DB.GetAllChirps(r.Context())
	if err != nil {
		log.Printf("Error fetching chirps: %v", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	sort.Slice(chirps, func(i, j int) bool {
		if sortOrder == "desc" {
			return chirps[i].CreatedAt.After(chirps[j].CreatedAt)
		}

		// default: asc
		return chirps[i].CreatedAt.Before(chirps[j].CreatedAt)
	})

	respondJSON(w, http.StatusOK, chirps)
}

func (h *APIHandler) GetMyChirps(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	sortOrder := query.Get("sort")

	if sortOrder == "" {
		sortOrder = "asc"
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Access token absent: %v", err)
		errJSON(w, http.StatusUnauthorized, ErrMessage{
			Message: "Unauthorized",
		})
		return
	}

	userID, err := auth.ValidateJWT(token, h.cfg.JWTSecret)
	if err != nil {
		log.Printf("Invalid access token: %v", err)
		errJSON(w, http.StatusUnauthorized, ErrMessage{
			Message: "Unauthorized",
		})
		return
	}

	chirps, err := h.cfg.DB.GetChirpsByUser(r.Context(), userID)
	if err != nil {
		log.Printf("Error fetching chirps: %v", err)
		http.Error(w, "Something went wrong", http.StatusNotFound)
		return
	}

	sort.Slice(chirps, func(i, j int) bool {
		if sortOrder == "desc" {
			return chirps[i].CreatedAt.After(chirps[j].CreatedAt)
		}
		// default: asc
		return chirps[i].CreatedAt.Before(chirps[j].CreatedAt)
	})

	respondJSON(w, http.StatusOK, chirps)
}

func (h *APIHandler) GetChirpsByUser(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	sortOrder := query.Get("sort")

	if sortOrder == "" {
		sortOrder = "asc"
	}

	userID, err := uuid.Parse(r.PathValue("userID"))
	if err != nil {
		log.Printf("Could not parse user ID: %v", err)
		errJSON(w, http.StatusBadRequest, ErrMessage{
			Message: "Something went wrong",
		})
		return
	}

	chirps, err := h.cfg.DB.GetChirpsByUser(r.Context(), userID)
	if err != nil {
		log.Printf("Error fetching chirps: %v", err)
		http.Error(w, "Something went wrong", http.StatusNotFound)
		return
	}

	sort.Slice(chirps, func(i, j int) bool {
		if sortOrder == "desc" {
			return chirps[i].CreatedAt.After(chirps[j].CreatedAt)
		}
		// default: asc
		return chirps[i].CreatedAt.Before(chirps[j].CreatedAt)
	})

	respondJSON(w, http.StatusOK, chirps)
}

func (h *APIHandler) GetChirpByChirpID(w http.ResponseWriter, r *http.Request) {
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		log.Printf("Could not parse chirp ID: %v", err)
		errJSON(w, http.StatusBadRequest, ErrMessage{
			Message: "Something went wrong",
		})
		return
	}
	chirp, err := h.cfg.DB.GetChirp(r.Context(), chirpID)
	if err != nil {
		log.Printf("Error fetching chirps: %v", err)
		http.Error(w, "Something went wrong", http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, chirp)
}

func (h *APIHandler) UpdateChirp(w http.ResponseWriter, r *http.Request) {
	chirpIDStr := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(chirpIDStr)
	if err != nil {
		errJSON(w, http.StatusBadRequest, ErrMessage{Message: "Invalid chirp ID"})
		return
	}

	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Access token absent: %v", err)
		errJSON(w, http.StatusUnauthorized, ErrMessage{
			Message: "Unauthorized",
		})
		return
	}

	userID, err := auth.ValidateJWT(accessToken, h.cfg.JWTSecret)
	if err != nil {
		log.Printf("Invalid access token: %v", err)
		errJSON(w, http.StatusUnauthorized, ErrMessage{
			Message: "Unauthorized",
		})
		return
	}

	user, err := h.cfg.DB.GetUserByID(r.Context(), userID)
	if err != nil {
		log.Printf("Error fetching user: %v", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	if !user.IsChirpyRed {
		log.Print("User not allowed to edit chirp")
		errJSON(w, http.StatusForbidden, ErrMessage{
			Message: "Only Chirpy Red members can edit chirps",
		})
		return
	}

	diff := ChirpBody{}
	err = json.NewDecoder(r.Body).Decode(&diff)
	if err != nil {
		log.Printf("Error decoding requested JSON: %v", err)
		errJSON(w, http.StatusBadRequest, ErrMessage{
			Message: "Something went wrong",
		})
		return
	}

	existing, err := h.cfg.DB.GetChirp(r.Context(), chirpID)
	if err != nil {
		errJSON(w, http.StatusNotFound, ErrMessage{Message: "Chirp not found"})
		return
	}

	if existing.UserID != userID {
		errJSON(w, http.StatusForbidden, ErrMessage{
			Message: "Not allowed to edit this chirp",
		})
		return
	}

	updatedChirp, err := h.cfg.DB.UpdateChirpBody(r.Context(), database.UpdateChirpBodyParams{
		Body: diff.Body,
		ID:   chirpID,
	})
	if err != nil {
		log.Printf("Error updating chirp: %v", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, updatedChirp)
}

func (h *APIHandler) DeleteChirp(w http.ResponseWriter, r *http.Request) {
	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Access token absent: %v", err)
		errJSON(w, http.StatusUnauthorized, ErrMessage{
			Message: "Unauthorized",
		})
		return
	}

	userID, err := auth.ValidateJWT(accessToken, h.cfg.JWTSecret)
	if err != nil {
		log.Printf("Invalid access token: %v", err)
		errJSON(w, http.StatusUnauthorized, ErrMessage{
			Message: "Unauthorized",
		})
		return
	}

	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		log.Printf("Could not parse chirp ID: %v", err)
		errJSON(w, http.StatusNotFound, ErrMessage{
			Message: "Chirp not found",
		})
		return
	}

	chirp, err := h.cfg.DB.GetChirp(r.Context(), chirpID)
	if err != nil {
		log.Printf("Error fetching chirp: %v", err)
		http.Error(w, "Something went wrong", http.StatusNotFound)
		return
	}

	if chirp.UserID != userID {
		log.Print("Unauthorized user cannot delete chirp")
		http.Error(w, "Unauthorized User", http.StatusForbidden)
		return
	}

	err = h.cfg.DB.DeleteChirp(r.Context(), chirpID)
	if err != nil {
		log.Printf("Error deleting chirp: %v", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// utility:
func cleanProfanity(text string) string {
	words := strings.Split(text, " ")
	for i, word := range words {
		var leading strings.Builder
		trailing := ""
		core := word

		for len(core) > 0 && isPunctuation(rune(core[0])) {
			leading.WriteString(string(core[0]))
			core = core[1:]
		}

		for len(core) > 0 && isPunctuation(rune(core[len(core)-1])) {
			trailing = string(core[len(core)-1]) + trailing
			core = core[:len(core)-1]
		}

		if slices.Contains(profane, strings.ToLower(core)) {
			words[i] = leading.String() + asterisk + trailing
		}
	}
	return strings.Join(words, " ")
}

func isPunctuation(r rune) bool {
	return strings.ContainsRune(".,!?;:'\"", r)
}
