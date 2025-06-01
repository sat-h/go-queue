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

	queue := job.NewQueue(os.Getenv("REDIS_ADDR"))
	processor := &job.SimpleProcessor{}
	w := &worker.Worker{Queue: queue, Processor: processor}

	log.Println("Worker started...")
	w.Start(ctx)
}
