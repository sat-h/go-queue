package worker

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/redis/go-redis/v9"
	"github.com/sat-h/go-queue/internal/job"
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
			if err == redis.Nil {
				time.Sleep(time.Second)
				continue
			}
			if err != nil {
				log.Printf("Error dequeuing: %v", err)
				continue
			}

			var j job.Job
			if err := json.Unmarshal([]byte(result), &j); err != nil {
				log.Printf("Invalid job: %v", err)
				continue
			}

			jobs <- j
		}
	}
}

func (w *Worker) handleJob(ctx context.Context, j job.Job) {
	operation := func() error {
		return w.Processor.Process(ctx, j)
	}

	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.MaxElapsedTime = 30 * time.Second

	if err := backoff.Retry(operation, expBackoff); err != nil {
		log.Printf("Job failed after retries: %v", err)
		// Optional: move to dead-letter queue here
	}
}
