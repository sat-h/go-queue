package job

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSimpleProcessor_Process(t *testing.T) {
	cases := []struct {
		name        string
		job         Job
		ctxTimeout  time.Duration
		wantErr     bool
		minDuration time.Duration
		maxDuration time.Duration
	}{
		{
			name:        "success, not cancelled",
			job:         Job{ID: "1", Payload: "test"},
			ctxTimeout:  3 * time.Second,
			wantErr:     false,
			minDuration: 2 * time.Second,
			maxDuration: 3 * time.Second,
		},
		{
			name:        "cancelled before finish",
			job:         Job{ID: "2", Payload: "cancel"},
			ctxTimeout:  500 * time.Millisecond,
			wantErr:     true,
			minDuration: 500,
			maxDuration: 2 * time.Second,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := &SimpleProcessor{}
			ctx, cancel := context.WithTimeout(context.Background(), tc.ctxTimeout)
			defer cancel()

			start := time.Now()
			err := p.Process(ctx, tc.job)
			duration := time.Since(start)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.GreaterOrEqual(t, duration, tc.minDuration)
			assert.LessOrEqual(t, duration, tc.maxDuration)
		})
	}
}
