package job

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

type Queue struct {
	client *redis.Client
	key    string
}

func NewQueue(redisAddr string) *Queue {
	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	return &Queue{client: client, key: "jobs"}
}

func (q *Queue) Enqueue(ctx context.Context, j Job) error {
	data, err := json.Marshal(j)
	if err != nil {
		return err
	}
	return q.client.RPush(ctx, q.key, data).Err()
}
