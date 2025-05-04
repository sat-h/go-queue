package job

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

//LEARN: Using the job.Queue wrapper provides encapsulation:
//Redis becomes an internal implementation detail, so you can switch it out later without modifying the handler.

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

func (q *Queue) Dequeue(ctx context.Context) (string, error) {
	return q.client.LPop(ctx, q.key).Result()
}

func (q *Queue) PrintAllJobs(ctx context.Context) error {
	results, err := q.client.LRange(ctx, q.key, 0, -1).Result()
	if err != nil {
		return err
	}

	for i, item := range results {
		var j Job
		if err := json.Unmarshal([]byte(item), &j); err != nil {
			fmt.Printf("Invalid job at index %d: %v\n", i, err)
			continue
		}
		fmt.Printf("Job #%d: %+v\n", i+1, j)
	}

	return nil
}
