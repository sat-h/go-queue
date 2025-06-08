# Redis Connection Troubleshooting Session

## Problem
When running the go-queue-worker in Kubernetes, it was showing connection issues:
```
Error dequeuing: dial tcp [::1]:6379: connect: connection refused
```

Later, when running the API service, we encountered similar issues:
```
curl -X POST http://localhost:8090/jobs -H "Content-Type: application/json" -d '{"payload": "test_job"}'
failed to enqueue job
```

## Root Cause
The worker code was attempting to connect to Redis twice:
1. First trying localhost ([::1]:6379)
2. Then trying the correct service address (redis:6379)

This happened because of how the Redis connection address was being built from environment variables.

The API service had a similar issue but was only using the REDIS_ADDR environment variable without proper fallback to REDIS_HOST and REDIS_PORT.

Additionally, in Kubernetes environments, the short service name "redis" sometimes had DNS resolution issues, requiring the use of fully-qualified domain names (FQDNs).

## Solution
### 1. Worker Code Fix
The worker code was modified to properly handle the Redis address determination with clear precedence between environment variables:

```go
// Determine Redis address with proper environment variable precedence
var redisAddr string

// First check REDIS_ADDR env var
redisAddr = os.Getenv("REDIS_ADDR")

// If not set, build from host and port
if redisAddr == "" {
    host := os.Getenv("REDIS_HOST")
    port := os.Getenv("REDIS_PORT")
    if host == "" {
        host = "localhost" // Default to localhost if REDIS_HOST not set
    }
    if port == "" {
        port = "6379" // Default to 6379 if REDIS_PORT not set
    }
    redisAddr = host + ":" + port
}

log.Printf("Connecting to Redis at: %s", redisAddr)
```

### 2. API Code Fix
Similar improvements were made to the API service:

```go
// Determine Redis address with proper environment variable precedence
var redisAddr string

// First check REDIS_ADDR env var
redisAddr = os.Getenv("REDIS_ADDR")

// If not set, build from host and port
if redisAddr == "" {
    host := os.Getenv("REDIS_HOST")
    port := os.Getenv("REDIS_PORT")
    if host == "" {
        host = "localhost" // Default to localhost if REDIS_HOST not set
    }
    if port == "" {
        port = "6379" // Default to 6379 if REDIS_PORT not set
    }
    redisAddr = host + ":" + port
}

log.Printf("API connecting to Redis at: %s", redisAddr)
```

### 3. Enhanced Redis Queue Implementation
The Queue implementation was improved with better error handling and logging:

```go
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
```

### 4. Kubernetes Configuration Fix
The final fix involved modifying the API deployment in `k8s/base/api.yaml` to use fully-qualified domain names for Redis:

```yaml
env:
- name: REDIS_ADDR
  value: "redis.go-queue.svc.cluster.local:6379"  # Explicitly set FQDN
- name: REDIS_HOST
  value: "redis.go-queue.svc.cluster.local"  # Using FQDN instead of just 'redis'
- name: REDIS_PORT
  value: "6379"
```

## Deployment Process
1. Rebuilt the worker container with the updated code:
   ```bash
   docker build -t go-queue-worker:latest -f docker/Dockerfile.worker .
   ```

2. Loaded the image into Minikube:
   ```bash
   minikube image load go-queue-worker:latest
   ```

3. Restarted the deployment to use the new image:
   ```bash
   kubectl rollout restart deployment/go-queue-worker -n go-queue
   ```

4. Applied the same process for the API service:
   ```bash
   docker build -t go-queue-api:latest -f docker/Dockerfile.api .
   minikube image load go-queue-api:latest
   kubectl apply -f k8s/base/api.yaml
   ```

5. Verified the deployment:
   ```bash
   kubectl get pods -n go-queue
   ```

6. Confirmed the Redis connection was successful by testing the API:
   ```bash
   curl -X POST http://localhost:8092/jobs -H "Content-Type: application/json" -d '{"payload": "test_job"}'
   ```

## Key Lessons Learned
1. In Kubernetes environments, use fully-qualified domain names (FQDNs) for service-to-service communication: `servicename.namespace.svc.cluster.local`
2. Always implement proper fallback mechanisms for environment variables
3. Add connection validation and meaningful error messages to quickly identify issues
4. Use appropriate timeouts for network operations
