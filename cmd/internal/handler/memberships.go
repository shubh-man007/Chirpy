package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/shubh-man007/Chirpy/cmd/internal/auth"
)

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
