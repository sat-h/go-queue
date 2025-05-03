package api

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/sat-h/go-queue/internal/job"
)

type Handler struct {
	Queue *job.Queue
}

func (h *Handler) SubmitJob(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Payload string `json:"payload"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	j := job.Job{
		ID:      uuid.New().String(),
		Payload: payload.Payload,
	}

	ctx := r.Context()
	if err := h.Queue.Enqueue(ctx, j); err != nil {
		http.Error(w, "failed to enqueue job", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(j)
}
