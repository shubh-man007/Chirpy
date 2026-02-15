// DB Model:
// Follower: The user who follows; Followee: The user who is followed

package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/shubh-man007/Chirpy/cmd/internal/auth"
	"github.com/shubh-man007/Chirpy/cmd/internal/database"
)

func (h *APIHandler) FollowUser(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		errJSON(w, http.StatusUnauthorized, ErrMessage{Message: "Unauthorized"})
		return
	}

	followerID, err := auth.ValidateJWT(token, h.cfg.JWTSecret)
	if err != nil {
		errJSON(w, http.StatusUnauthorized, ErrMessage{Message: "Unauthorized"})
		return
	}

	var req struct {
		FolloweeID string `json:"followee_id"`
	}

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		errJSON(w, http.StatusBadRequest, ErrMessage{Message: "Invalid Request"})
		return
	}

	followeeID, err := uuid.Parse(req.FolloweeID)
	if err != nil {
		errJSON(w, http.StatusBadRequest, ErrMessage{Message: "Invalid user ID format"})
		return
	}

	if followerID == followeeID {
		errJSON(w, http.StatusBadRequest, ErrMessage{
			Message: "Cannot follow self",
		})
		return
	}

	err = h.cfg.DB.FollowUser(r.Context(), database.FollowUserParams{
		FollowerID: followerID,
		FolloweeID: followeeID,
	})
	if err != nil {
		log.Printf("Error following user: %v", err)
		errJSON(w, http.StatusInternalServerError, ErrMessage{Message: "Failed to follow user"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *APIHandler) UnfollowUser(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		errJSON(w, http.StatusUnauthorized, ErrMessage{Message: "Unauthorized"})
		return
	}

	followerID, err := auth.ValidateJWT(token, h.cfg.JWTSecret)
	if err != nil {
		errJSON(w, http.StatusUnauthorized, ErrMessage{Message: "Unauthorized"})
		return
	}

	followeeIDStr := r.PathValue("userID")
	followeeID, err := uuid.Parse(followeeIDStr)
	if err != nil {
		errJSON(w, http.StatusBadRequest, ErrMessage{Message: "Invalid user ID format"})
		return
	}

	if followerID == followeeID {
		errJSON(w, http.StatusBadRequest, ErrMessage{
			Message: "Cannot unfollow self",
		})
		return
	}

	err = h.cfg.DB.UnfollowUser(r.Context(), database.UnfollowUserParams{
		FollowerID: followerID,
		FolloweeID: followeeID,
	})
	if err != nil {
		log.Printf("Error unfollowing user: %v", err)
		errJSON(w, http.StatusInternalServerError, ErrMessage{Message: "Failed to unfollow user"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *APIHandler) GetFollowers(w http.ResponseWriter, r *http.Request) {
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

	followers, err := h.cfg.DB.GetFollowers(r.Context(), userID)
	if err != nil {
		log.Printf("Error fetching followers: %v", err)
		errJSON(w, http.StatusInternalServerError, ErrMessage{Message: "Failed to fetch followers"})
		return
	}

	var followerCount int64
	if len(followers) > 0 {
		followerCount = followers[0].TotalFollowers
	}

	type FollowerInfo struct {
		Followers     []database.GetFollowersRow
		FollowerCount int64
	}

	respondJSON(w, http.StatusOK, FollowerInfo{
		Followers:     followers,
		FollowerCount: followerCount,
	})
}

func (h *APIHandler) GetFollowing(w http.ResponseWriter, r *http.Request) {
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

	following, err := h.cfg.DB.GetFollowing(r.Context(), userID)
	if err != nil {
		log.Printf("Error fetching following users: %v", err)
		errJSON(w, http.StatusInternalServerError, ErrMessage{Message: "Failed to fetch following users"})
		return
	}

	var followingCount int64
	if len(following) > 0 {
		followingCount = following[0].TotalFollowing
	}

	type FollowingInfo struct {
		Following      []database.GetFollowingRow
		FollowingCount int64
	}

	respondJSON(w, http.StatusOK, FollowingInfo{
		Following:      following,
		FollowingCount: followingCount,
	})
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

	query := r.URL.Query()
	limitStr := query.Get("limit")
	offsetStr := query.Get("offset")

	var limit int32
	if limitStr == "" {
		limit = 20
	} else {
		val, _ := strconv.Atoi(limitStr)
		limit = int32(val)
	}

	var offset int32
	if offsetStr == "" {
		offset = 20
	} else {
		val, _ := strconv.Atoi(offsetStr)
		offset = int32(val)
	}

	chirps, err := h.cfg.DB.GetFeed(r.Context(), database.GetFeedParams{
		FollowerID: userID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		log.Printf("Error fetching feed: %v", err)
		errJSON(w, http.StatusInternalServerError, ErrMessage{Message: "Failed to fetch feed"})
		return
	}

	respondJSON(w, http.StatusOK, chirps)
}
