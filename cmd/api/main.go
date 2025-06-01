package main

import (
	"log"
	"net/http"
	"os"

	"github.com/sat-h/go-queue/internal/api"
	"github.com/sat-h/go-queue/internal/job"
)

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func readinessHandler(w http.ResponseWriter, _ *http.Request) {
	// Optionally, add deeper checks (e.g., Redis connectivity)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ready"))
}

func main() {
	queue := job.NewQueue(os.Getenv("REDIS_ADDR"))
	handler := &api.Handler{Queue: queue}
	router := api.NewRouter(handler)

	mux := http.NewServeMux()
	mux.Handle("/healthz", http.HandlerFunc(healthHandler))
	mux.Handle("/readyz", http.HandlerFunc(readinessHandler))
	mux.Handle("/", router)

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("API running on :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
