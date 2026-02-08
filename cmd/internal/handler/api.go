package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"sort"
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

// users
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
	err := json.NewDecoder(r.Body).Decode(&updateCred)
	if err != nil {
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

func (h *APIHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	pathUserID, err := uuid.Parse(r.PathValue("userID"))
	if err != nil {
		log.Printf("Could not parse user ID: %v", err)
		errJSON(w, http.StatusBadRequest, ErrMessage{
			Message: "Something went wrong",
		})
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

	if pathUserID != userID {
		log.Print("UserID not matched with ID sent via path")
		errJSON(w, http.StatusForbidden, ErrMessage{
			Message: "Unauthorized",
		})
		return
	}

	err = h.cfg.DB.DeleteUserByID(r.Context(), userID)
	if err != nil {
		log.Printf("Error deleting user: %v", err)
		http.Error(w, "Something went wrong", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// chirps
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

// Membership
func (h *APIHandler) UpdateUserMembership(w http.ResponseWriter, r *http.Request) {
	webhookEvent := MembershipWebhookEvent{}
	err := json.NewDecoder(r.Body).Decode(&webhookEvent)
	if err != nil {
		log.Printf("Error decoding webhook event: %v", err)
		errJSON(w, http.StatusBadRequest, ErrMessage{
			Message: "Something went wrong",
		})
		return
	}

	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		log.Printf("API Key absent: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if apiKey != h.cfg.PolkaAPIKey {
		log.Printf("Invalid API Key: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	event := webhookEvent.Event
	if event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	userID, err := uuid.Parse(webhookEvent.Data.UserID)
	if err != nil {
		log.Printf("Could not parse userID in webhook event: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.cfg.DB.UpdateUserToChirpyRed(r.Context(), userID)
	if err != nil {
		log.Printf("User not found: %v", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// friends
func (h *APIHandler) SendFriendRequest(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		errJSON(w, http.StatusUnauthorized, ErrMessage{Message: "Unauthorized"})
		return
	}

	userID, err := auth.ValidateJWT(token, h.cfg.JWTSecret)
	if err != nil {
		errJSON(w, http.StatusUnauthorized, ErrMessage{Message: "Unauthorized"})
		return
	}

	var req struct {
		FriendID string `json:"friend_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errJSON(w, http.StatusBadRequest, ErrMessage{Message: "Invalid request"})
		return
	}

	friendID, err := uuid.Parse(req.FriendID)
	if err != nil {
		errJSON(w, http.StatusBadRequest, ErrMessage{Message: "Invalid user ID"})
		return
	}

	if userID == friendID {
		errJSON(w, http.StatusBadRequest, ErrMessage{Message: "Cannot send friend request to yourself"})
		return
	}

	friendship, err := h.cfg.DB.CreateFriendRequest(r.Context(), database.CreateFriendRequestParams{
		UserID:   userID,
		FriendID: friendID,
	})
	if err != nil {
		log.Printf("Error creating friend request: %v", err)
		errJSON(w, http.StatusInternalServerError, ErrMessage{Message: "Failed to send friend request"})
		return
	}

	respondJSON(w, http.StatusCreated, friendship)
}

func (h *APIHandler) AcceptFriendRequest(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		errJSON(w, http.StatusUnauthorized, ErrMessage{Message: "Unauthorized"})
		return
	}

	userID, err := auth.ValidateJWT(token, h.cfg.JWTSecret)
	if err != nil {
		errJSON(w, http.StatusUnauthorized, ErrMessage{Message: "Unauthorized"})
		return
	}

	friendIDStr := r.PathValue("userID")
	friendID, err := uuid.Parse(friendIDStr)
	if err != nil {
		errJSON(w, http.StatusBadRequest, ErrMessage{Message: "Invalid user ID"})
		return
	}

	err = h.cfg.DB.AcceptFriendRequest(r.Context(), database.AcceptFriendRequestParams{
		UserID:   userID,
		FriendID: friendID,
	})
	if err != nil {
		log.Printf("Error accepting friend request: %v", err)
		errJSON(w, http.StatusInternalServerError, ErrMessage{Message: "Failed to accept friend request"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *APIHandler) RejectFriendRequest(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		errJSON(w, http.StatusUnauthorized, ErrMessage{Message: "Unauthorized"})
		return
	}

	userID, err := auth.ValidateJWT(token, h.cfg.JWTSecret)
	if err != nil {
		errJSON(w, http.StatusUnauthorized, ErrMessage{Message: "Unauthorized"})
		return
	}

	friendIDStr := r.PathValue("userID")
	friendID, err := uuid.Parse(friendIDStr)
	if err != nil {
		errJSON(w, http.StatusBadRequest, ErrMessage{Message: "Invalid user ID"})
		return
	}

	err = h.cfg.DB.RejectFriendRequest(r.Context(), database.RejectFriendRequestParams{
		UserID:   userID,
		FriendID: friendID,
	})
	if err != nil {
		log.Printf("Error rejecting friend request: %v", err)
		errJSON(w, http.StatusInternalServerError, ErrMessage{Message: "Failed to reject friend request"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *APIHandler) RemoveFriend(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		errJSON(w, http.StatusUnauthorized, ErrMessage{Message: "Unauthorized"})
		return
	}

	userID, err := auth.ValidateJWT(token, h.cfg.JWTSecret)
	if err != nil {
		errJSON(w, http.StatusUnauthorized, ErrMessage{Message: "Unauthorized"})
		return
	}

	friendIDStr := r.PathValue("userID")
	friendID, err := uuid.Parse(friendIDStr)
	if err != nil {
		errJSON(w, http.StatusBadRequest, ErrMessage{Message: "Invalid user ID"})
		return
	}

	err = h.cfg.DB.RemoveFriendship(r.Context(), database.RemoveFriendshipParams{
		UserID:   userID,
		FriendID: friendID,
	})
	if err != nil {
		log.Printf("Error removing friendship: %v", err)
		errJSON(w, http.StatusInternalServerError, ErrMessage{Message: "Failed to remove friendship"})
		return
	}

	w.WriteHeader(http.StatusNoContent)

}

func (h *APIHandler) GetFriends(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		errJSON(w, http.StatusUnauthorized, ErrMessage{Message: "Unauthorized"})
		return
	}

	userID, err := auth.ValidateJWT(token, h.cfg.JWTSecret)
	if err != nil {
		errJSON(w, http.StatusUnauthorized, ErrMessage{Message: "Unauthorized"})
		return
	}

	friends, err := h.cfg.DB.GetFriends(r.Context(), userID)
	if err != nil {
		log.Printf("Error fetching friends: %v", err)
		http.Error(w, "Something went wrong", http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, friends)
}

func (h *APIHandler) GetPendingFriendRequests(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		errJSON(w, http.StatusUnauthorized, ErrMessage{Message: "Unauthorized"})
		return
	}

	userID, err := auth.ValidateJWT(token, h.cfg.JWTSecret)
	if err != nil {
		errJSON(w, http.StatusUnauthorized, ErrMessage{Message: "Unauthorized"})
		return
	}

	pendingFriends, err := h.cfg.DB.GetPendingRequests(r.Context(), userID)
	if err != nil {
		log.Printf("Error fetching friends: %v", err)
		http.Error(w, "Something went wrong", http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, pendingFriends)
}

func (h *APIHandler) GetSentFriendRequests(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		errJSON(w, http.StatusUnauthorized, ErrMessage{Message: "Unauthorized"})
		return
	}

	userID, err := auth.ValidateJWT(token, h.cfg.JWTSecret)
	if err != nil {
		errJSON(w, http.StatusUnauthorized, ErrMessage{Message: "Unauthorized"})
		return
	}

	sentFriends, err := h.cfg.DB.GetSentRequests(r.Context(), userID)
	if err != nil {
		log.Printf("Error fetching friends: %v", err)
		http.Error(w, "Something went wrong", http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, sentFriends)
}

func (h *APIHandler) GetFeed(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		errJSON(w, http.StatusUnauthorized, ErrMessage{Message: "Unauthorized"})
		return
	}

	userID, err := auth.ValidateJWT(token, h.cfg.JWTSecret)
	if err != nil {
		errJSON(w, http.StatusUnauthorized, ErrMessage{Message: "Unauthorized"})
		return
	}

	limit := int32(20)
	offset := int32(0)

	chirps, err := h.cfg.DB.GetFriendFeed(r.Context(), database.GetFriendFeedParams{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		log.Printf("Error fetching feed: %v", err)
		errJSON(w, http.StatusInternalServerError, ErrMessage{Message: "Failed to fetch feed"})
		return
	}

	respondJSON(w, http.StatusOK, chirps)
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
