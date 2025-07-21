package worker

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/redis/go-redis/v9"
	"github.com/sat-h/go-queue/internal/job"
	"github.com/sat-h/go-queue/internal/observability/metrics"
)

type Worker struct {
	Queue     *job.Queue
	Processor job.Processor
}

func (w *Worker) Start(ctx context.Context) {
	jobs := make(chan job.Job)

	go w.poll(ctx, jobs)

	for {
		select {
		case j := <-jobs:
			go w.handleJob(ctx, j)
		case <-ctx.Done():
			log.Println("Worker shutting down")
			return
		}
	}
}

func (w *Worker) poll(ctx context.Context, jobs chan<- job.Job) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			result, err := w.Queue.Dequeue(ctx)
			if errors.Is(err, redis.Nil) {
				time.Sleep(time.Second)
				continue
			}
			if err != nil {
				log.Printf("Error dequeuing: %v", err)
				time.Sleep(time.Second) // Add backoff on error
				continue
			}

			var j job.Job
			if err := json.Unmarshal([]byte(result), &j); err != nil {
				log.Printf("Invalid job: %v", err)
				continue
			}

			// Check if this job has already been processed (deduplication)
			if j.ID != "" && w.Queue.IsProcessed(j.ID) {
				log.Printf("Skipping already processed job ID: %s", j.ID)
				// Count deduplication events in metrics
				metrics.JobProcessed("deduplicated", j.Type)
				continue
			}

			jobs <- j
		}
	}
}

func (w *Worker) handleJob(ctx context.Context, j job.Job) {
	startTime := time.Now()

	operation := func() error {
		return w.Processor.Process(ctx, j)
	}

	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.MaxElapsedTime = 30 * time.Second

	if err := backoff.Retry(operation, expBackoff); err != nil {
		log.Printf("Job failed after retries: %v", err)
		metrics.JobProcessed("failed", j.Type)
		return
	}

	// Mark job as processed to avoid duplicates during failover
	if j.ID != "" {
		w.Queue.MarkProcessed(j.ID)
	}

	// Record metrics for successful processing
	duration := time.Since(startTime).Seconds()
	metrics.ObserveJobProcessingTime(j.Type, duration)
	metrics.JobProcessed("success", j.Type)
}
