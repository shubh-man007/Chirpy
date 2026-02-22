package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/shubh-man007/Chirpy/cmd/internal/auth"
	"github.com/shubh-man007/Chirpy/cmd/internal/database"
)

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

func (h *APIHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(r.PathValue("userID"))
	if err != nil {
		log.Printf("Could not parse user ID: %v", err)
		errJSON(w, http.StatusBadRequest, ErrMessage{
			Message: "Something went wrong",
		})
		return
	}

	user, err := h.cfg.DB.GetUserByID(r.Context(), userID)
	if err != nil {
		log.Printf("Error fetching user: %v", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, user)
}
