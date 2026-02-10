package handler

// import (
// 	"database/sql"
// 	"encoding/json"
// 	"errors"
// 	"log"
// 	"net/http"

// 	"github.com/google/uuid"
// 	"github.com/shubh-man007/Chirpy/cmd/internal/auth"
// 	"github.com/shubh-man007/Chirpy/cmd/internal/database"
// )

// func (h *APIHandler) SendFriendRequest(w http.ResponseWriter, r *http.Request) {
// 	token, err := auth.GetBearerToken(r.Header)
// 	if err != nil {
// 		errJSON(w, http.StatusUnauthorized, ErrMessage{Message: "Unauthorized"})
// 		return
// 	}

// 	userID, err := auth.ValidateJWT(token, h.cfg.JWTSecret)
// 	if err != nil {
// 		errJSON(w, http.StatusUnauthorized, ErrMessage{Message: "Unauthorized"})
// 		return
// 	}

// 	var req struct {
// 		FriendID string `json:"friend_id"`
// 	}
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		errJSON(w, http.StatusBadRequest, ErrMessage{Message: "Invalid request"})
// 		return
// 	}

// 	friendID, err := uuid.Parse(req.FriendID)
// 	if err != nil {
// 		errJSON(w, http.StatusBadRequest, ErrMessage{Message: "Invalid user ID format"})
// 		return
// 	}

// 	if userID == friendID {
// 		errJSON(w, http.StatusBadRequest, ErrMessage{Message: "Cannot send friend request to yourself"})
// 		return
// 	}

// 	_, err = h.cfg.DB.GetUserByID(r.Context(), friendID)
// 	if err != nil {
// 		log.Printf("Friend user not found: %v", err)
// 		errJSON(w, http.StatusNotFound, ErrMessage{Message: "User not found"})
// 		return
// 	}

// 	areFriends, err := h.cfg.DB.AreFriends(r.Context(), database.AreFriendsParams{
// 		UserID:   userID,
// 		FriendID: friendID,
// 	})
// 	if err == nil && areFriends {
// 		errJSON(w, http.StatusBadRequest, ErrMessage{Message: "Already friends with this user"})
// 		return
// 	}

// 	status, err := h.cfg.DB.GetFriendshipStatus(r.Context(), database.GetFriendshipStatusParams{
// 		UserID:   userID,
// 		FriendID: friendID,
// 	})
// 	if err == nil && status == "pending" {
// 		errJSON(w, http.StatusBadRequest, ErrMessage{Message: "Friend request already sent"})
// 		return
// 	}

// 	friendship, err := h.cfg.DB.CreateFriendRequest(r.Context(), database.CreateFriendRequestParams{
// 		UserID:   userID,
// 		FriendID: friendID,
// 	})
// 	if err != nil {
// 		log.Printf("Error creating friend request: %v", err)
// 		errJSON(w, http.StatusInternalServerError, ErrMessage{Message: "Failed to send friend request"})
// 		return
// 	}

// 	respondJSON(w, http.StatusCreated, friendship)
// }

// func (h *APIHandler) AcceptFriendRequest(w http.ResponseWriter, r *http.Request) {
// 	token, err := auth.GetBearerToken(r.Header)
// 	if err != nil {
// 		errJSON(w, http.StatusUnauthorized, ErrMessage{Message: "Unauthorized"})
// 		return
// 	}

// 	userID, err := auth.ValidateJWT(token, h.cfg.JWTSecret)
// 	if err != nil {
// 		errJSON(w, http.StatusUnauthorized, ErrMessage{Message: "Unauthorized"})
// 		return
// 	}

// 	friendIDStr := r.PathValue("userID")
// 	friendID, err := uuid.Parse(friendIDStr)
// 	if err != nil {
// 		errJSON(w, http.StatusBadRequest, ErrMessage{Message: "Invalid user ID"})
// 		return
// 	}

// 	err = h.cfg.DB.AcceptFriendRequestSafely(r.Context(), database.AcceptFriendRequestParams{
// 		UserID:   userID,
// 		FriendID: friendID,
// 	})
// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			errJSON(w, http.StatusNotFound, ErrMessage{
// 				Message: "No pending friend request from this user",
// 			})
// 			return
// 		} else {
// 			log.Printf("Error accepting friend request: %v", err)
// 			errJSON(w, http.StatusInternalServerError, ErrMessage{Message: "Failed to accept friend request"})
// 			return
// 		}
// 	}
// 	w.WriteHeader(http.StatusNoContent)
// }

// func (h *APIHandler) RejectFriendRequest(w http.ResponseWriter, r *http.Request) {
// 	token, err := auth.GetBearerToken(r.Header)
// 	if err != nil {
// 		errJSON(w, http.StatusUnauthorized, ErrMessage{Message: "Unauthorized"})
// 		return
// 	}

// 	userID, err := auth.ValidateJWT(token, h.cfg.JWTSecret)
// 	if err != nil {
// 		errJSON(w, http.StatusUnauthorized, ErrMessage{Message: "Unauthorized"})
// 		return
// 	}

// 	friendIDStr := r.PathValue("userID")
// 	friendID, err := uuid.Parse(friendIDStr)
// 	if err != nil {
// 		errJSON(w, http.StatusBadRequest, ErrMessage{Message: "Invalid user ID"})
// 		return
// 	}

// 	err = h.cfg.DB.RejectFriendRequestSafely(r.Context(), database.RejectFriendRequestParams{
// 		UserID:   userID,
// 		FriendID: friendID,
// 	})
// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			errJSON(w, http.StatusNotFound, ErrMessage{
// 				Message: "No pending friend request from this user",
// 			})
// 			return
// 		} else {
// 			log.Printf("Error rejecting friend request: %v", err)
// 			errJSON(w, http.StatusInternalServerError, ErrMessage{Message: "Failed to reject friend request"})
// 			return
// 		}
// 	}

// 	w.WriteHeader(http.StatusNoContent)
// }

// func (h *APIHandler) RemoveFriend(w http.ResponseWriter, r *http.Request) {
// 	token, err := auth.GetBearerToken(r.Header)
// 	if err != nil {
// 		errJSON(w, http.StatusUnauthorized, ErrMessage{Message: "Unauthorized"})
// 		return
// 	}

// 	userID, err := auth.ValidateJWT(token, h.cfg.JWTSecret)
// 	if err != nil {
// 		errJSON(w, http.StatusUnauthorized, ErrMessage{Message: "Unauthorized"})
// 		return
// 	}

// 	friendIDStr := r.PathValue("userID")
// 	friendID, err := uuid.Parse(friendIDStr)
// 	if err != nil {
// 		errJSON(w, http.StatusBadRequest, ErrMessage{Message: "Invalid user ID"})
// 		return
// 	}

// 	err = h.cfg.DB.RemoveFriendship(r.Context(), database.RemoveFriendshipParams{
// 		UserID:   userID,
// 		FriendID: friendID,
// 	})
// 	if err != nil {
// 		log.Printf("Error removing friendship: %v", err)
// 		errJSON(w, http.StatusInternalServerError, ErrMessage{Message: "Failed to remove friendship"})
// 		return
// 	}

// 	w.WriteHeader(http.StatusNoContent)

// }

// func (h *APIHandler) GetFriends(w http.ResponseWriter, r *http.Request) {
// 	token, err := auth.GetBearerToken(r.Header)
// 	if err != nil {
// 		errJSON(w, http.StatusUnauthorized, ErrMessage{Message: "Unauthorized"})
// 		return
// 	}

// 	userID, err := auth.ValidateJWT(token, h.cfg.JWTSecret)
// 	if err != nil {
// 		errJSON(w, http.StatusUnauthorized, ErrMessage{Message: "Unauthorized"})
// 		return
// 	}

// 	friends, err := h.cfg.DB.GetFriends(r.Context(), userID)
// 	if err != nil {
// 		log.Printf("Error fetching friends: %v", err)
// 		http.Error(w, "Something went wrong", http.StatusNotFound)
// 		return
// 	}

// 	respondJSON(w, http.StatusOK, friends)
// }

// func (h *APIHandler) GetPendingFriendRequests(w http.ResponseWriter, r *http.Request) {
// 	token, err := auth.GetBearerToken(r.Header)
// 	if err != nil {
// 		errJSON(w, http.StatusUnauthorized, ErrMessage{Message: "Unauthorized"})
// 		return
// 	}

// 	userID, err := auth.ValidateJWT(token, h.cfg.JWTSecret)
// 	if err != nil {
// 		errJSON(w, http.StatusUnauthorized, ErrMessage{Message: "Unauthorized"})
// 		return
// 	}

// 	pendingFriends, err := h.cfg.DB.GetPendingRequests(r.Context(), userID)
// 	if err != nil {
// 		log.Printf("Error fetching friends: %v", err)
// 		http.Error(w, "Something went wrong", http.StatusNotFound)
// 		return
// 	}

// 	respondJSON(w, http.StatusOK, pendingFriends)
// }

// func (h *APIHandler) GetSentFriendRequests(w http.ResponseWriter, r *http.Request) {
// 	token, err := auth.GetBearerToken(r.Header)
// 	if err != nil {
// 		errJSON(w, http.StatusUnauthorized, ErrMessage{Message: "Unauthorized"})
// 		return
// 	}

// 	userID, err := auth.ValidateJWT(token, h.cfg.JWTSecret)
// 	if err != nil {
// 		errJSON(w, http.StatusUnauthorized, ErrMessage{Message: "Unauthorized"})
// 		return
// 	}

// 	sentFriends, err := h.cfg.DB.GetSentRequests(r.Context(), userID)
// 	if err != nil {
// 		log.Printf("Error fetching friends: %v", err)
// 		http.Error(w, "Something went wrong", http.StatusNotFound)
// 		return
// 	}

// 	respondJSON(w, http.StatusOK, sentFriends)
// }

// func (h *APIHandler) GetFeed(w http.ResponseWriter, r *http.Request) {
// 	token, err := auth.GetBearerToken(r.Header)
// 	if err != nil {
// 		errJSON(w, http.StatusUnauthorized, ErrMessage{Message: "Unauthorized"})
// 		return
// 	}

// 	userID, err := auth.ValidateJWT(token, h.cfg.JWTSecret)
// 	if err != nil {
// 		errJSON(w, http.StatusUnauthorized, ErrMessage{Message: "Unauthorized"})
// 		return
// 	}

// 	limit := int32(20)
// 	offset := int32(0)

// 	chirps, err := h.cfg.DB.GetFriendFeed(r.Context(), database.GetFriendFeedParams{
// 		UserID: userID,
// 		Limit:  limit,
// 		Offset: offset,
// 	})
// 	if err != nil {
// 		log.Printf("Error fetching feed: %v", err)
// 		errJSON(w, http.StatusInternalServerError, ErrMessage{Message: "Failed to fetch feed"})
// 		return
// 	}

// 	respondJSON(w, http.StatusOK, chirps)
// }
