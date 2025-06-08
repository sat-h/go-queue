package api

import (
	"encoding/json"
	"fmt"
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
		log.Printf("Error decoding payload: %v", err)
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	j := job.Job{
		ID:      uuid.New().String(),
		Payload: payload.Payload,
	}

	ctx := r.Context()
	log.Printf("Attempting to enqueue job with ID: %s", j.ID)
	if err := h.Queue.Enqueue(ctx, j); err != nil {
		log.Printf("Error enqueuing job: %v", err)
		errorMsg := fmt.Sprintf("failed to enqueue job: %v", err)
		http.Error(w, errorMsg, http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully enqueued job %s", j.ID)
	if err := h.Queue.PrintAllJobs(ctx); err != nil {
		log.Printf("Error printing jobs: %v", err)
	}
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(j)
}
