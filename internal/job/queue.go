package job

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

//LEARN: Using the job.Queue wrapper provides encapsulation:
//Redis becomes an internal implementation detail, so you can switch it out later without modifying the handler.

type Queue struct {
	client *redis.Client
	key    string
}

func NewQueue(redisAddr string) *Queue {
	log.Printf("Creating Redis client with address: %s", redisAddr)
	client := redis.NewClient(&redis.Options{
		Addr:        redisAddr,
		DialTimeout: 5 * time.Second, // Add reasonable timeout
	})

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Printf("WARNING: Failed to connect to Redis at %s: %v", redisAddr, err)
	} else {
		log.Printf("Successfully connected to Redis at %s", redisAddr)
	}

	return &Queue{client: client, key: "jobs"}
}

func (q *Queue) Enqueue(ctx context.Context, j Job) error {
	data, err := json.Marshal(j)
	if err != nil {
		return fmt.Errorf("error marshaling job: %w", err)
	}

	err = q.client.RPush(ctx, q.key, data).Err()
	if err != nil {
		return fmt.Errorf("Redis RPush error: %w", err)
	}

	return nil
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
