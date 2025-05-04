package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/sat-h/go-queue/internal/job"
	"github.com/sat-h/go-queue/internal/worker"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// LEARN: job.NewQueue("localhost:6379") is called in worker/main.go and api/main.go. It returns a redis client not the shared queue.
	queue := job.NewQueue("localhost:6379")
	processor := &job.SimpleProcessor{}
	w := &worker.Worker{Queue: queue, Processor: processor}

	log.Println("Worker started...")
	w.Start(ctx)
}
