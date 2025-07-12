# Phase 7: High Availability & Fault Tolerance Implementation Guide

This guide provides a step-by-step approach to implementing high availability and fault tolerance features for your Go queue system. It focuses on Redis Sentinel setup, job deduplication, and additional metrics for better observability during failure scenarios.

## Table of Contents
1. [Redis HA with Sentinel Implementation](#1-redis-ha-with-sentinel-implementation)
2. [Job Deduplication Implementation](#2-job-deduplication-implementation)
3. [HA-Specific Metrics Implementation](#3-ha-specific-metrics-implementation)
4. [Kubernetes Configuration for HA](#4-kubernetes-configuration-for-ha)
5. [Testing and Validation](#5-testing-and-validation)
6. [Post MVP Tasks](#6-post-mvp-tasks)

## 1. Redis HA with Sentinel Implementation

### Step 1.1: Create Redis HA Kubernetes Configuration

Create a new file `k8s/base/redis-ha.yaml` with the following content:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-config
  namespace: go-queue
data:
  master.conf: |
    port 6379
    dir /data
    appendonly yes
    protected-mode no
  replica.conf: |
    port 6379
    dir /data
    appendonly yes
    protected-mode no
    replicaof redis-master 6379
  sentinel.conf: |
    port 26379
    dir /tmp
    sentinel monitor mymaster redis-master 6379 1
    sentinel down-after-milliseconds mymaster 5000
    sentinel failover-timeout mymaster 60000
    sentinel parallel-syncs mymaster 1
    sentinel announce-ip $(POD_IP)
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: redis-master
  namespace: go-queue
spec:
  serviceName: redis-master
  replicas: 1
  selector:
    matchLabels:
      app: redis-master
  template:
    metadata:
      labels:
        app: redis-master
    spec:
      containers:
      - name: redis
        image: redis:6.2-alpine
        ports:
        - containerPort: 6379
          name: redis
        volumeMounts:
        - name: config
          mountPath: /usr/local/etc/redis/redis.conf
          subPath: master.conf
        - name: data
          mountPath: /data
        command: ["redis-server", "/usr/local/etc/redis/redis.conf"]
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 200m
            memory: 256Mi
      volumes:
      - name: config
        configMap:
          name: redis-config
      - name: data
        emptyDir: {}
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: redis-replica
  namespace: go-queue
spec:
  serviceName: redis-replica
  replicas: 1
  selector:
    matchLabels:
      app: redis-replica
  template:
    metadata:
      labels:
        app: redis-replica
    spec:
      containers:
      - name: redis
        image: redis:6.2-alpine
        ports:
        - containerPort: 6379
          name: redis
        volumeMounts:
        - name: config
          mountPath: /usr/local/etc/redis/redis.conf
          subPath: replica.conf
        - name: data
          mountPath: /data
        command: ["redis-server", "/usr/local/etc/redis/redis.conf"]
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 200m
            memory: 256Mi
      volumes:
      - name: config
        configMap:
          name: redis-config
      - name: data
        emptyDir: {}
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: redis-sentinel
  namespace: go-queue
spec:
  serviceName: redis-sentinel
  replicas: 1
  selector:
    matchLabels:
      app: redis-sentinel
  template:
    metadata:
      labels:
        app: redis-sentinel
    spec:
      containers:
      - name: sentinel
        image: redis:6.2-alpine
        env:
        - name: POD_IP
          valueFrom:
            fieldRef:
              fieldPath: status.podIP
        ports:
        - containerPort: 26379
          name: sentinel
        volumeMounts:
        - name: config
          mountPath: /usr/local/etc/redis/sentinel.conf
          subPath: sentinel.conf
        command: ["redis-sentinel", "/usr/local/etc/redis/sentinel.conf"]
        resources:
          requests:
            cpu: 50m
            memory: 64Mi
          limits:
            cpu: 100m
            memory: 128Mi
      volumes:
      - name: config
        configMap:
          name: redis-config
---
apiVersion: v1
kind: Service
metadata:
  name: redis-master
  namespace: go-queue
spec:
  selector:
    app: redis-master
  ports:
  - port: 6379
    targetPort: 6379
    name: redis
  clusterIP: None
---
apiVersion: v1
kind: Service
metadata:
  name: redis-replica
  namespace: go-queue
spec:
  selector:
    app: redis-replica
  ports:
  - port: 6379
    targetPort: 6379
    name: redis
  clusterIP: None
---
apiVersion: v1
kind: Service
metadata:
  name: redis-sentinel
  namespace: go-queue
spec:
  selector:
    app: redis-sentinel
  ports:
  - port: 26379
    targetPort: 26379
    name: sentinel
  clusterIP: None
```

### Step 1.2: Update kustomization.yaml

Update your `k8s/base/kustomization.yaml` file to include the new Redis HA configuration:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- namespace.yaml
- configmap.yaml
- secrets.yaml
- redis-ha.yaml  # Use this instead of redis.yaml
- api.yaml
- worker.yaml
- ingress.yaml
```

## 2. Job Deduplication Implementation

### Step 2.1: Create Job ID Map for Deduplication

Modify `internal/job/queue.go` to add job deduplication:

1. First, add a new struct to track processed job IDs:

```go
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
```

### Step 2.2: Modify Queue for Sentinel Support and Deduplication

2. Next, update the Queue struct and add configuration options:

```go
type Queue struct {
	client       redis.UniversalClient
	key          string
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
	} else {
		log.Printf("Successfully connected to Redis")
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
		client:       client, 
		key:          "jobs",
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
```

### Step 2.3: Add Deduplication Methods to Queue

3. Add methods to mark and check processed jobs:

```go
// Add these new methods to the Queue struct

func (q *Queue) MarkProcessed(id string) {
	q.processedJobs.Add(id)
}

func (q *Queue) IsProcessed(id string) bool {
	return q.processedJobs.Exists(id)
}

// Update Enqueue to ensure job has an ID
func (q *Queue) Enqueue(ctx context.Context, j Job) error {
	// Ensure job has an ID for deduplication
	if j.ID == "" {
		return fmt.Errorf("job must have an ID")
	}
	
	// ... existing marshaling and Redis code ...
}
```

### Step 2.4: Update Worker to Use Deduplication

Modify `internal/worker/worker.go` to check for already processed jobs:

```go
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
				// Count deduplication events in metrics (added in next section)
				metrics.JobProcessed("deduplicated", "unknown")
				continue
			}

			jobs <- j
		}
	}
}

func (w *Worker) handleJob(ctx context.Context, j job.Job) {
	startTime := time.Now()
	// ... existing retry/processing code ...

	if err := backoff.Retry(operation, expBackoff); err != nil {
		log.Printf("Job failed after retries: %v", err)
		metrics.JobProcessed("failed", "unknown")
		return
	}
	
	// Mark job as processed to avoid duplicates during failover
	if j.ID != "" {
		w.Queue.MarkProcessed(j.ID)
	}
	
	// ... existing metrics code ...
}
```

## 3. HA-Specific Metrics Implementation

### Step 3.1: Add New HA Metrics

Update `internal/observability/metrics/metrics.go` to add failover and HA-specific metrics:

```go
package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Existing metrics
	jobsProcessed *prometheus.CounterVec
	jobProcessingTime *prometheus.HistogramVec
	queueLength prometheus.Gauge
	
	// New HA-related metrics
	redisConnectionFailures *prometheus.CounterVec
	redisReconnectionAttempts prometheus.Counter
	redisReconnectionSuccess prometheus.Counter
	jobDeduplicationEvents *prometheus.CounterVec
	workerRecoveryTime *prometheus.HistogramVec
	
	once sync.Once
)

// Init initializes the metrics collection system
func Init() {
	once.Do(func() {
		// Existing metrics initialization
		jobsProcessed = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "jobs_processed_total",
				Help: "The total number of processed jobs",
			},
			[]string{"status", "job_type"},
		)

		jobProcessingTime = promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "job_processing_duration_seconds",
				Help:    "Time taken to process jobs",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"job_type"},
		)

		queueLength = promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "queue_length",
				Help: "Current number of jobs in the queue",
			},
		)
		
		// New HA metrics
		redisConnectionFailures = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "redis_connection_failures_total",
				Help: "Total number of Redis connection failures",
			},
			[]string{"connection_type"}, // "sentinel", "master", "replica"
		)
		
		redisReconnectionAttempts = promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "redis_reconnection_attempts_total",
				Help: "Total number of Redis reconnection attempts",
			},
		)
		
		redisReconnectionSuccess = promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "redis_reconnection_success_total",
				Help: "Total number of successful Redis reconnections",
			},
		)
		
		jobDeduplicationEvents = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "job_deduplication_events_total",
				Help: "Total number of prevented duplicate job processing events",
			},
			[]string{"job_type"},
		)
		
		workerRecoveryTime = promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "worker_recovery_time_seconds",
				Help: "Time taken for worker to recover after failures",
				Buckets: prometheus.ExponentialBuckets(0.1, 2, 10), // 0.1s to ~100s
			},
			[]string{"failure_type"}, // "redis", "processing", etc.
		)
	})
}

// Existing metrics methods

// New HA metrics methods
func RecordRedisConnectionFailure(connectionType string) {
	if redisConnectionFailures != nil {
		redisConnectionFailures.WithLabelValues(connectionType).Inc()
	}
}

func RecordRedisReconnectionAttempt() {
	if redisReconnectionAttempts != nil {
		redisReconnectionAttempts.Inc()
	}
}

func RecordRedisReconnectionSuccess() {
	if redisReconnectionSuccess != nil {
		redisReconnectionSuccess.Inc()
	}
}

func RecordJobDeduplication(jobType string) {
	if jobDeduplicationEvents != nil {
		jobDeduplicationEvents.WithLabelValues(jobType).Inc()
	}
}

func ObserveWorkerRecoveryTime(failureType string, durationSeconds float64) {
	if workerRecoveryTime != nil {
		workerRecoveryTime.WithLabelValues(failureType).Observe(durationSeconds)
	}
}

// Getter methods for testing
func GetRedisConnectionFailuresCounter() *prometheus.CounterVec {
	return redisConnectionFailures
}

func GetRedisReconnectionAttemptsCounter() prometheus.Counter {
	return redisReconnectionAttempts
}

func GetRedisReconnectionSuccessCounter() prometheus.Counter {
	return redisReconnectionSuccess
}

func GetJobDeduplicationEventsCounter() *prometheus.CounterVec {
	return jobDeduplicationEvents
}

func GetWorkerRecoveryTimeHistogram() *prometheus.HistogramVec {
	return workerRecoveryTime
}
```

### Step 3.2: Update Queue to Record Metrics

Modify the `Queue` in `internal/job/queue.go` to record reconnection metrics:

```go
// Update the NewQueueWithOptions function to include metrics recording
func NewQueueWithOptions(opts QueueOptions) *Queue {
	var client redis.UniversalClient
	// ...existing code...

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), opts.DialTimeout)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Printf("WARNING: Failed to connect to Redis: %v", err)
		metrics.RecordRedisConnectionFailure(strings.ToLower(opts.RedisMode))
	} else {
		log.Printf("Successfully connected to Redis")
	}
	
	// ...existing code...
}
```

### Step 3.3: Update Worker to Record Deduplication Metrics

In `internal/worker/worker.go`, update the polling function to record deduplication events:

```go
// In poll method
// Check if this job has already been processed (deduplication)
if j.ID != "" && w.Queue.IsProcessed(j.ID) {
    log.Printf("Skipping already processed job ID: %s", j.ID)
    // Record deduplication event
    metrics.RecordJobDeduplication(j.Type) // If job has Type field, otherwise use "unknown"
    metrics.JobProcessed("deduplicated", "unknown")
    continue
}
```

## 4. Kubernetes Configuration for HA

### Step 4.1: Update API Deployment

Modify `k8s/base/api.yaml` to use multiple replicas and sentinel configuration:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api
  namespace: go-queue
spec:
  replicas: 2  # Multiple replicas for high availability
  selector:
    matchLabels:
      app: api
  template:
    metadata:
      labels:
        app: api
    spec:
      containers:
      - name: api
        image: go-queue-api:latest
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 8080
        envFrom:
        - configMapRef:
            name: go-queue-config
        - secretRef:
            name: go-queue-secrets
        env:
        - name: REDIS_MODE
          value: "sentinel"
        - name: REDIS_SENTINEL_ADDRS
          value: "redis-sentinel.go-queue.svc.cluster.local:26379"
        - name: REDIS_MASTER_NAME
          value: "mymaster"
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 200m
            memory: 256Mi
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 15
        readinessProbe:
          httpGet:
            path: /readyz
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
```

### Step 4.2: Update Worker Deployment

Modify `k8s/base/worker.yaml` similarly:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: go-queue-worker
  namespace: go-queue
spec:
  replicas: 2  # Multiple replicas for high availability
  selector:
    matchLabels:
      app: worker
  template:
    metadata:
      labels:
        app: worker
    spec:
      containers:
      - name: worker
        image: go-queue-worker:latest
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 8081  # For health and metrics endpoints
        envFrom:
        - configMapRef:
            name: go-queue-config
        - secretRef:
            name: go-queue-secrets
        env:
        - name: REDIS_MODE
          value: "sentinel"
        - name: REDIS_SENTINEL_ADDRS
          value: "redis-sentinel.go-queue.svc.cluster.local:26379"
        - name: REDIS_MASTER_NAME
          value: "mymaster"
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 200m
            memory: 256Mi
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8081
          initialDelaySeconds: 10
          periodSeconds: 15
        readinessProbe:
          httpGet:
            path: /readyz
            port: 8081
          initialDelaySeconds: 5
          periodSeconds: 10
```

### Step 4.3: Update Main Functions for Sentinel Support

Update `cmd/api/main.go` to support Redis Sentinel:

```go
// After observability initialization, replace Redis connection code with:

// Configure Redis connection based on environment variables
var queue *job.Queue

// Check if we're using Redis Sentinel
redisMode := os.Getenv("REDIS_MODE")
if redisMode == "" {
    redisMode = "standalone" // Default to standalone mode
}

logger.Info("Configuring Redis client", zap.String("mode", redisMode))

switch strings.ToLower(redisMode) {
case "sentinel":
    // Get Sentinel configuration
    sentinelAddrs := os.Getenv("REDIS_SENTINEL_ADDRS")
    if sentinelAddrs == "" {
        sentinelAddrs = "redis-sentinel:26379" // Default to our k8s service
    }
    
    masterName := os.Getenv("REDIS_MASTER_NAME")
    if masterName == "" {
        masterName = "mymaster" // Default master name from our sentinel config
    }
    
    // Split the sentinel addresses
    sentinels := strings.Split(sentinelAddrs, ",")
    
    logger.Info("Connecting to Redis via Sentinel", 
        zap.Strings("sentinels", sentinels),
        zap.String("master", masterName))
    
    queue = job.NewQueueWithOptions(job.QueueOptions{
        RedisMode:   "sentinel",
        RedisAddrs:  sentinels,
        MasterName:  masterName,
        Password:    os.Getenv("REDIS_PASSWORD"),
    })
    
default:
    // Determine Redis address with proper environment variable precedence
    var redisAddr string

    // First check REDIS_ADDR env var
    redisAddr = os.Getenv("REDIS_ADDR")

    // If not set, build from host and port
    if redisAddr == "" {
        host := os.Getenv("REDIS_HOST")
        port := os.Getenv("REDIS_PORT")
        if host == "" {
            // Use fully-qualified Kubernetes DNS name instead of just "redis"
            host = "redis.go-queue.svc.cluster.local"
            logger.Info("REDIS_HOST not set, using Kubernetes FQDN", zap.String("host", host))
        }
        if port == "" {
            port = "6379" // Default to 6379 if REDIS_PORT not set
        }
        redisAddr = host + ":" + port
    }

    logger.Info("API connecting to Redis in standalone mode", zap.String("address", redisAddr))
    queue = job.NewQueue(redisAddr)
}

handler := &api.Handler{Queue: queue}
// ... rest of the code ...
```

Update `cmd/worker/main.go` similarly with the same Redis connection code.

## 5. Testing and Validation

### Step 5.1: Create Testing Guide

Create a file `docs/high_availability_testing_guide.md` with comprehensive test scenarios:

```markdown
# High Availability Testing Guide

This guide outlines manual test procedures for validating the high availability and fault tolerance of the go-queue system.

## Prerequisites

- Minikube cluster running
- kubectl CLI configured
- go-queue system deployed with Redis HA configuration

## Test Scenario 1: Redis Master Failure

### Objective
Validate that the system continues to function when the Redis master fails, with Sentinel promoting a replica to master.

### Procedure

1. **Create some test jobs**

```bash
# Port-forward the API service
kubectl port-forward -n go-queue svc/api 8080:8080

# In another terminal, submit 10 test jobs
for i in {1..10}; do
    curl -X POST -H "Content-Type: application/json" \
    -d "{\"id\":\"test-$i\",\"payload\":\"This is test job $i\"}" \
    http://localhost:8080/api/v1/jobs
done
```

2. **Verify jobs are being processed**

```bash
# View worker logs to confirm jobs are being processed
kubectl logs -n go-queue -l app=worker --tail=20
```

3. **Identify and kill the Redis master pod**

```bash
# First, get Redis pod names
kubectl get pods -n go-queue -l app=redis-master

# Kill the master pod 
kubectl delete pod -n go-queue redis-master-0
```

4. **Observe failover**

```bash
# Watch Redis Sentinel logs to observe master promotion
kubectl logs -n go-queue -l app=redis-sentinel -f

# Monitor worker logs to see connection handling
kubectl logs -n go-queue -l app=worker -f
```

5. **Submit new jobs during/after failover**

```bash
# Submit 5 more jobs
for i in {11..15}; do
    curl -X POST -H "Content-Type: application/json" \
    -d "{\"id\":\"test-$i\",\"payload\":\"This is test job $i\"}" \
    http://localhost:8080/api/v1/jobs
done
```

6. **Document observations**
   - How long was the system unavailable?
   - Were any jobs lost or duplicated? 
   - Did the workers reconnect automatically?
   - Did Sentinel promote the replica successfully?

## Test Scenario 2: API Pod Failure

### Objective
Ensure the system continues to accept jobs when an API pod fails, with traffic redirected to the remaining healthy pod.

### Procedure

1. **Identify API pods**

```bash
kubectl get pods -n go-queue -l app=api
```

2. **Kill one of the API pods**

```bash
kubectl delete pod -n go-queue <api-pod-name>
```

3. **Immediately attempt to submit jobs**

```bash
# Submit jobs while the pod is being replaced
for i in {1..5}; do
    curl -X POST -H "Content-Type: application/json" \
    -d "{\"id\":\"api-test-$i\",\"payload\":\"API failover test job $i\"}" \
    http://localhost:8080/api/v1/jobs
done
```

4. **Check Prometheus metrics**

```bash
# Port-forward Prometheus service (if available)
kubectl port-forward -n monitoring svc/prometheus-server 9090:9090
```

Then visit http://localhost:9090 in your browser and query for:
- `redis_connection_failures_total`
- `job_deduplication_events_total`

## Test Scenario 3: Worker Pod Failure

### Objective
Validate that job processing continues when a worker pod fails, with no job loss or duplication.

### Procedure

1. **Submit a batch of jobs with unique IDs**

```bash
for i in {1..20}; do
    curl -X POST -H "Content-Type: application/json" \
    -d "{\"id\":\"worker-test-$i\",\"payload\":\"Worker failover test job $i\"}" \
    http://localhost:8080/api/v1/jobs
done
```

2. **While jobs are processing, kill one worker pod**

```bash
kubectl get pods -n go-queue -l app=worker
kubectl delete pod -n go-queue <worker-pod-name>
```

3. **Observe the remaining worker and the new worker pod**

```bash
kubectl logs -n go-queue -l app=worker -f
```

4. **Document observations**
   - Was job processing interrupted?
   - Were any jobs processed twice (check for deduplication logs)?
   - How long did it take the new worker pod to start processing jobs?

## Documentation Template

```
Test Scenario: [Name]
Date/Time: [When test was conducted]
------------------------------------
Configuration: 
- Redis: [Sentinel config details]
- API Replicas: [Number]
- Worker Replicas: [Number]

Observations:
- Downtime: [Duration in seconds]
- Recovery time: [Duration in seconds]
- Job statistics: [Processed/Duplicated/Lost]
- System behavior: [Describe what happened]

Metrics:
- [Any relevant metrics from Prometheus]

Conclusion:
- [Overall assessment of HA behavior]
```
```

### Step 5.2: Deploy and Validate

1. Build and deploy your updated application:

```bash
# Apply the updated Kubernetes configuration
kubectl apply -k k8s/base/

# Verify the deployment
kubectl get pods -n go-queue
```

2. Follow the testing guide to validate the HA capabilities.

## 6. Post MVP Tasks

The current implementation guide focuses on the MVP version of high availability with Redis Sentinel (1 master, 1 replica, 1 Sentinel). For future improvements after the MVP phase, the following Post MVP tasks are identified:

### 6.1 Post MVP 1: Proper Sentinel HA (1 master, 2 replicas, 3 Sentinels)

For true high availability, the system should be upgraded to:
- 1 Redis master
- 2 Redis replicas (instead of just one)
- 3 Redis Sentinels (instead of just one)

This configuration provides:
- Real failover with quorum-based decision making
- Higher redundancy and automatic recovery
- More reliable failure detection

Key modifications would involve:
- Updating the `redis-ha.yaml` to increase replica and sentinel replicas to appropriate counts
- Adjusting the Sentinel configuration for proper quorum settings
- Testing with more complex failure scenarios

### 6.2 Post MVP 2: Managed Redis or Redis Cluster

For production-grade high availability, scaling, and operational simplicity:
- Consider using a managed Redis service (AWS ElastiCache, Google Cloud Memorystore, Azure Cache for Redis)
- Alternatively, implement Redis Cluster for horizontal scaling and sharding

Benefits include:
- Automatic failover, backups, and scaling
- Reduced operational overhead
- Built-in monitoring and alerting
- Horizontal scalability for higher throughput

Note that managed solutions may incur costs and require cloud resources, but provide the most robust HA solution for production environments.

---

This guide provides a comprehensive approach to implementing high availability and fault tolerance features for your Go queue system. By implementing Redis Sentinel for HA, job deduplication to prevent duplicate processing, and additional metrics for better observability, your system will be able to recover automatically from failures while maintaining data integrity.
