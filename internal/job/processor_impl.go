package job

import (
	"context"
	"log"
	"time"
)

type ProcessorImpl struct{}

func (p *ProcessorImpl) Process(ctx context.Context, job Job) error {
	log.Printf("Processing job: %+v", job)
	// Simulate work
	select {
	case <-time.After(2 * time.Second):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
