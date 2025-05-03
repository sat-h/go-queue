package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter(h *Handler) http.Handler {
	r := chi.NewRouter()
	r.Post("/jobs", h.SubmitJob)
	return r
}
