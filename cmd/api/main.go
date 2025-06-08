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
	// Determine Redis address with proper environment variable precedence
	var redisAddr string

	// First check REDIS_ADDR env var
	redisAddr = os.Getenv("REDIS_ADDR")

	// If not set, build from host and port
	if redisAddr == "" {
		host := os.Getenv("REDIS_HOST")
		port := os.Getenv("REDIS_PORT")
		if host == "" {
			// Use fully-qualified Kubernetes DNS name instead of just "redis"
			host = "redis.go-queue.svc.cluster.local"
			log.Printf("REDIS_HOST not set, using Kubernetes FQDN: %s", host)
		}
		if port == "" {
			port = "6379" // Default to 6379 if REDIS_PORT not set
		}
		redisAddr = host + ":" + port
	}

	log.Printf("API connecting to Redis at: %s", redisAddr)
	queue := job.NewQueue(redisAddr)
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
