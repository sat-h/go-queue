package job

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sat-h/go-queue/internal/observability/metrics"
)

//LEARN: Using the job.Queue wrapper provides encapsulation:
//Redis becomes an internal implementation detail, so you can switch it out later without modifying the handler.

// JobIDMap is used to track processed job IDs for deduplication
type JobIDMap struct {
	ids map[string]time.Time
	mu  sync.Mutex
}

// NewJobIDMap creates a new JobIDMap
func NewJobIDMap() *JobIDMap {
	return &JobIDMap{
		ids: make(map[string]time.Time),
	}
}

// Add adds a job ID to the map with the current time
func (m *JobIDMap) Add(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ids[id] = time.Now()
}

// Exists checks if a job ID exists in the map
func (m *JobIDMap) Exists(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, exists := m.ids[id]
	return exists
}

// Cleanup removes job IDs older than the specified duration
func (m *JobIDMap) Cleanup(olderThan time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cutoff := time.Now().Add(-olderThan)
	for id, t := range m.ids {
		if t.Before(cutoff) {
			delete(m.ids, id)
		}
	}
}

type Queue struct {
	client        redis.UniversalClient
	key           string
	processedJobs *JobIDMap
}

// Configuration options for creating a new Queue
type QueueOptions struct {
	RedisAddrs   []string // Can be single Redis address or multiple for Sentinel
	RedisMode    string   // "standalone", "sentinel", or "cluster"
	MasterName   string   // Used only for sentinel mode
	Password     string   // Redis password if any
	DB           int      // Redis database number
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	DedupWindow  time.Duration // How long to track processed jobs for deduplication
}

// NewQueueWithOptions creates a new Queue with detailed configuration options
func NewQueueWithOptions(opts QueueOptions) *Queue {
	var client redis.UniversalClient

	// Set default values if not provided
	if opts.DialTimeout == 0 {
		opts.DialTimeout = 5 * time.Second
	}
	if opts.ReadTimeout == 0 {
		opts.ReadTimeout = 3 * time.Second
	}
	if opts.WriteTimeout == 0 {
		opts.WriteTimeout = 3 * time.Second
	}
	if opts.DedupWindow == 0 {
		opts.DedupWindow = 24 * time.Hour
	}

	switch strings.ToLower(opts.RedisMode) {
	case "sentinel":
		log.Printf("Creating Redis client with Sentinel. Master: %s, Sentinels: %v",
			opts.MasterName, opts.RedisAddrs)
		client = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:    opts.MasterName,
			SentinelAddrs: opts.RedisAddrs,
			Password:      opts.Password,
			DB:            opts.DB,
			DialTimeout:   opts.DialTimeout,
			ReadTimeout:   opts.ReadTimeout,
			WriteTimeout:  opts.WriteTimeout,
		})
	case "cluster":
		log.Printf("Creating Redis client in cluster mode. Addresses: %v", opts.RedisAddrs)
		client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        opts.RedisAddrs,
			Password:     opts.Password,
			DialTimeout:  opts.DialTimeout,
			ReadTimeout:  opts.ReadTimeout,
			WriteTimeout: opts.WriteTimeout,
		})
	default: // standalone
		log.Printf("Creating Redis client in standalone mode. Address: %s", opts.RedisAddrs[0])
		client = redis.NewClient(&redis.Options{
			Addr:         opts.RedisAddrs[0],
			Password:     opts.Password,
			DB:           opts.DB,
			DialTimeout:  opts.DialTimeout,
			ReadTimeout:  opts.ReadTimeout,
			WriteTimeout: opts.WriteTimeout,
		})
	}

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), opts.DialTimeout)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Printf("WARNING: Failed to connect to Redis: %v", err)
		metrics.RecordRedisConnectionFailure(strings.ToLower(opts.RedisMode))
	} else {
		log.Printf("Successfully connected to Redis")
		metrics.RecordRedisReconnectionSuccess()
	}

	processedJobs := NewJobIDMap()

	// Start a background goroutine to clean up the processed jobs map
	go func() {
		for {
			time.Sleep(1 * time.Hour)
			processedJobs.Cleanup(opts.DedupWindow)
		}
	}()

	return &Queue{
		client:        client,
		key:           "jobs",
		processedJobs: processedJobs,
	}
}

// Update existing NewQueue to use the new options
func NewQueue(redisAddr string) *Queue {
	return NewQueueWithOptions(QueueOptions{
		RedisAddrs: []string{redisAddr},
		RedisMode:  "standalone",
	})
}

// MarkProcessed adds a job ID to the processed map
func (q *Queue) MarkProcessed(id string) {
	q.processedJobs.Add(id)
}

// IsProcessed checks if a job ID exists in the processed map
func (q *Queue) IsProcessed(id string) bool {
	return q.processedJobs.Exists(id)
}

func (q *Queue) Enqueue(ctx context.Context, j Job) error {
	// Ensure job has an ID for deduplication
	if j.ID == "" {
		return fmt.Errorf("job must have an ID")
	}

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
