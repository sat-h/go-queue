package job

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
)

// TestQueue_Enqueue tests the Enqueue method for success, marshal error, and Redis error.
func TestQueue_Enqueue(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mockRedis, mock := redismock.NewClientMock()
	q := &Queue{client: mockRedis, key: "jobs"}

	job := Job{ID: "1", Payload: "test"}
	data, _ := json.Marshal(job)

	tests := []struct {
		name      string
		job       Job
		mockSetup func()
		wantErr   bool
	}{
		{
			name: "success",
			job:  job,
			mockSetup: func() {
				mock.ExpectRPush("jobs", data).SetVal(1)
			},
			wantErr: false,
		},
		{
			name: "marshal error",
			job:  Job{ID: "1", Payload: string([]byte{0xff, 0xfe})}, // invalid UTF-8
			mockSetup: func() {
				// No Redis call expected
			},
			wantErr: true,
		},
		{
			name: "redis error",
			job:  job,
			mockSetup: func() {
				mock.ExpectRPush("jobs", data).SetErr(errors.New("redis down"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.mockSetup()
			err := q.Enqueue(ctx, tt.job)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestQueue_Dequeue tests Dequeue for success, empty queue, and Redis error.
func TestQueue_Dequeue(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mockRedis, mock := redismock.NewClientMock()
	q := &Queue{client: mockRedis, key: "jobs"}

	tests := []struct {
		name      string
		mockSetup func()
		want      string
		wantErr   bool
	}{
		{
			name: "success",
			mockSetup: func() {
				mock.ExpectLPop("jobs").SetVal("jobdata")
			},
			want:    "jobdata",
			wantErr: false,
		},
		{
			name: "empty queue",
			mockSetup: func() {
				mock.ExpectLPop("jobs").RedisNil()
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "redis error",
			mockSetup: func() {
				mock.ExpectLPop("jobs").SetErr(errors.New("redis error"))
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.mockSetup()
			got, err := q.Dequeue(ctx)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestQueue_PrintAllJobs tests PrintAllJobs for all valid jobs, some invalid jobs, and Redis error.
func TestQueue_PrintAllJobs(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mockRedis, mock := redismock.NewClientMock()
	q := &Queue{client: mockRedis, key: "jobs"}

	validJob := Job{ID: "1", Payload: "test"}
	validData, _ := json.Marshal(validJob)
	invalidData := "not-json"

	tests := []struct {
		name      string
		mockSetup func()
		wantErr   bool
	}{
		{
			name: "all valid jobs",
			mockSetup: func() {
				mock.ExpectLRange("jobs", 0, -1).SetVal([]string{string(validData)})
			},
			wantErr: false,
		},
		{
			name: "some invalid jobs",
			mockSetup: func() {
				mock.ExpectLRange("jobs", 0, -1).SetVal([]string{string(validData), invalidData})
			},
			wantErr: false,
		},
		{
			name: "redis error",
			mockSetup: func() {
				mock.ExpectLRange("jobs", 0, -1).SetErr(errors.New("redis error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.mockSetup()
			err := q.PrintAllJobs(ctx)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
