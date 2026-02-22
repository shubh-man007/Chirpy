package handler

import (
	"log"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/shubh-man007/Chirpy/cmd/internal/auth"
	"github.com/shubh-man007/Chirpy/cmd/internal/database"
)

const defaultProfileChirpsLimit = 20
const maxProfileChirpsLimit = 100

func (h *APIHandler) buildProfileResponse(w http.ResponseWriter, r *http.Request, userID uuid.UUID, viewerID *uuid.UUID) {
	limit := int32(defaultProfileChirpsLimit)
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil && val > 0 {
			limit = int32(val)
			if limit > maxProfileChirpsLimit {
				limit = maxProfileChirpsLimit
			}
		}
	}

	var cursor uuid.NullUUID
	if cursorStr := r.URL.Query().Get("cursor"); cursorStr != "" {
		if cursorID, err := uuid.Parse(cursorStr); err == nil {
			cursor = uuid.NullUUID{UUID: cursorID, Valid: true}
		}
	}

	userStats, err := h.cfg.DB.GetUserProfile(r.Context(), userID)
	if err != nil {
		log.Printf("Error fetching profile: %v", err)
		errJSON(w, http.StatusNotFound, ErrMessage{Message: "User not found"})
		return
	}

	chirps, err := h.cfg.DB.GetUserChirpsPaginated(r.Context(), database.GetUserChirpsPaginatedParams{
		UserID:    userID,
		Cursor:    cursor,
		PageLimit: limit,
	})
	if err != nil {
		log.Printf("Error fetching chirps: %v", err)
		errJSON(w, http.StatusInternalServerError, ErrMessage{Message: "Failed to fetch chirps"})
		return
	}

	chirpItems := make([]ChirpItem, len(chirps))
	for i, c := range chirps {
		chirpItems[i] = ChirpItem{
			ID:        c.ID,
			Body:      c.Body,
			CreatedAt: c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	var nextCursor *string
	if len(chirps) == int(limit) {
		lastID := chirps[len(chirps)-1].ID.String()
		nextCursor = &lastID
	}

	profile := ProfileResponse{
		ID:             userStats.ID,
		Email:          userStats.Email,
		CreatedAt:      userStats.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:      userStats.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		IsChirpyRed:    userStats.IsChirpyRed,
		FollowersCount: userStats.FollowersCount,
		FollowingCount: userStats.FollowingCount,
		ChirpsCount:    userStats.ChirpsCount,
		Chirps:         chirpItems,
		NextCursor:     nextCursor,
	}

	if viewerID != nil && *viewerID != userID {
		following, err := h.cfg.DB.IsFollowing(r.Context(), database.IsFollowingParams{
			FollowerID: *viewerID,
			FolloweeID: userID,
		})
		if err == nil {
			profile.IsFollowing = &following
		}
	}

	respondJSON(w, http.StatusOK, profile)
}

func (h *APIHandler) GetMyProfile(w http.ResponseWriter, r *http.Request) {
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

	h.buildProfileResponse(w, r, userID, nil)
}

func (h *APIHandler) GetProfileByUserID(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(r.PathValue("userID"))
	if err != nil {
		errJSON(w, http.StatusBadRequest, ErrMessage{Message: "Invalid user ID"})
		return
	}

	var viewerID *uuid.UUID
	if token, err := auth.GetBearerToken(r.Header); err == nil {
		if id, err := auth.ValidateJWT(token, h.cfg.JWTSecret); err == nil {
			viewerID = &id
		}
	}

	h.buildProfileResponse(w, r, userID, viewerID)
}
