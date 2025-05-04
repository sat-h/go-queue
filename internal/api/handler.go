package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/sat-h/go-queue/internal/job"
)

// LEARN: Injecting a queue allows better testability (easy to mock)

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
	if err := h.Queue.PrintAllJobs(ctx); err != nil {
		log.Printf("print all jobs error %w", err)
	}
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(j)
}
