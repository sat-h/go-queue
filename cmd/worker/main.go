package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/sat-h/go-queue/internal/job"
	"github.com/sat-h/go-queue/internal/worker"
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
	// Start health/readiness endpoints in a goroutine
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/healthz", http.HandlerFunc(healthHandler))
		mux.Handle("/readyz", http.HandlerFunc(readinessHandler))

		port := os.Getenv("WORKER_PORT")
		if port == "" {
			port = "8081"
		}
		log.Printf("Worker health endpoints running on :%s\n", port)
		log.Fatal(http.ListenAndServe(":"+port, mux))
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Determine Redis address with proper environment variable precedence
	var redisAddr string

	// First check REDIS_ADDR env var
	redisAddr = os.Getenv("REDIS_ADDR")

	// If not set, build from host and port
	if redisAddr == "" {
		host := os.Getenv("REDIS_HOST")
		port := os.Getenv("REDIS_PORT")
		if host == "" {
			host = "localhost" // Default to localhost if REDIS_HOST not set
		}
		if port == "" {
			port = "6379" // Default to 6379 if REDIS_PORT not set
		}
		redisAddr = host + ":" + port
	}

	log.Printf("Connecting to Redis at: %s", redisAddr)
	queue := job.NewQueue(redisAddr)
	processor := &job.SimpleProcessor{}
	w := &worker.Worker{Queue: queue, Processor: processor}

	log.Println("Worker started...")
	w.Start(ctx)
}
