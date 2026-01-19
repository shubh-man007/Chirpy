package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"
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

func (h *APIHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req UserLogin
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request JSON: %v", err)
		errJSON(w, http.StatusBadRequest, ErrMessage{
			Message: "Invalid email or password",
		})
		return
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		errJSON(w, http.StatusInternalServerError, ErrMessage{
			Message: "Something went wrong",
		})
		return
	}

	userParams := database.CreateUserParams{
		Email:          req.Email,
		HashedPassword: hashedPassword,
	}
	user, err := h.cfg.DB.CreateUser(r.Context(), userParams)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusCreated, user)
}

func (h *APIHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var req UserLogin
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request JSON: %v", err)
		errJSON(w, http.StatusBadRequest, ErrMessage{
			Message: "Something went wrong",
		})
		return
	}

	userCreds, err := h.cfg.DB.GetUserPassByEmail(r.Context(), req.Email)
	if err != nil {
		log.Printf("Error validating user: %s", err)
		http.Error(w, "Invalid email or password", http.StatusInternalServerError)
		return
	}

	val, err := auth.CheckPasswordHash(req.Password, userCreds)
	if err != nil {
		log.Printf("Unauthorized User: %s", err)
		http.Error(w, "Something went wrong", http.StatusUnauthorized)
		return
	}

	if !val {
		log.Printf("Unauthorized User: %s", err)
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	user, err := h.cfg.DB.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		log.Printf("Error fetching user: %v", err)
		http.Error(w, "Invalid email or password", http.StatusInternalServerError)
		return
	}

	jwtToken, err := auth.MakeJWT(user.ID, h.cfg.JWTSecret, time.Hour)
	if err != nil {
		log.Printf("Error creating JWT token: %v", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		log.Printf("Error creating refresh token: %v", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	_, err = h.cfg.DB.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(60 * 24 * time.Hour), // 60 days
	})
	if err != nil {
		log.Printf("Error storing refresh token: %v", err)
		errJSON(w, http.StatusInternalServerError, ErrMessage{
			Message: "Something went wrong",
		})
		return
	}

	type LoginResponse struct {
		database.GetUserByEmailRow
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	respondJSON(w, http.StatusOK, LoginResponse{
		GetUserByEmailRow: user,
		Token:             jwtToken,
		RefreshToken:      refreshToken,
	})
}

func (h *APIHandler) UpdateUserCred(w http.ResponseWriter, r *http.Request) {
	var updateCred UserLogin
	if err := json.NewDecoder(r.Body).Decode(&updateCred); err != nil {
		log.Printf("Error decoding request JSON: %v", err)
		errJSON(w, http.StatusBadRequest, ErrMessage{
			Message: "Invalid email or password",
		})
		return
	}

	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Access token absent: %v", err)
		http.Error(w, "Something went wrong", http.StatusUnauthorized)
		return
	}

	userID, err := auth.ValidateJWT(accessToken, h.cfg.JWTSecret)
	if err != nil {
		log.Printf("Invalid access token: %v", err)
		http.Error(w, "Something went wrong", http.StatusUnauthorized)
		return
	}

	hashedPassword, err := auth.HashPassword(updateCred.Password)
	if err != nil {
		log.Printf("Could not hash password: %v", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	updatedUser, err := h.cfg.DB.UpdateUserCred(r.Context(), database.UpdateUserCredParams{
		Email:          updateCred.Email,
		HashedPassword: hashedPassword,
		ID:             userID,
	})
	if err != nil {
		log.Printf("Could not update DB: %v", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, updatedUser)
}

func (h *APIHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Refresh token absent: %v", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	user, err := h.cfg.DB.GetUserFromRefreshToken(r.Context(), refreshToken)
	if err != nil {
		log.Printf("Invalid refresh token: %v", err)
		errJSON(w, http.StatusUnauthorized, ErrMessage{
			Message: "Unauthorized",
		})
		return
	}

	accessToken, err := auth.MakeJWT(user.ID, h.cfg.JWTSecret, time.Hour)
	if err != nil {
		log.Printf("Error creating access token: %v", err)
		errJSON(w, http.StatusInternalServerError, ErrMessage{
			Message: "Something went wrong",
		})
		return
	}

	type RefreshResponse struct {
		Token string `json:"token"`
	}

	respondJSON(w, http.StatusOK, RefreshResponse{
		Token: accessToken,
	})
}

func (h *APIHandler) RevokeToken(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Refresh token absent: %v", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	err = h.cfg.DB.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		log.Printf("Error revoking token: %v", err)
		errJSON(w, http.StatusInternalServerError, ErrMessage{
			Message: "Something went wrong",
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *APIHandler) CreateChirp(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	chirp := ChirpBody{}
	err := decoder.Decode(&chirp)
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
	chirps, err := h.cfg.DB.GetAllChirps(r.Context())
	if err != nil {
		log.Printf("Error fetching chirps: %v", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, chirps)
}

func (h *APIHandler) GetChirpsByUser(w http.ResponseWriter, r *http.Request) {
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		log.Printf("Could not parse chirp ID: %v", err)
		errJSON(w, http.StatusBadRequest, ErrMessage{
			Message: "Something went wrong",
		})
		return
	}
	chirps, err := h.cfg.DB.GetChirpsByUser(r.Context(), chirpID)
	if err != nil {
		log.Printf("Error fetching chirps: %v", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
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

func (h *APIHandler) DeleteChirp(w http.ResponseWriter, r *http.Request) {
	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Accent token absent: %v", err)
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

	user, err := h.cfg.DB.GetChirp(r.Context(), chirpID)
	if err != nil {
		log.Printf("Error fetching chirp: %v", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	if user.ID != userID {
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
