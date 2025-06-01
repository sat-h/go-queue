resilient-job-queue/
├── cmd/
│   ├── api/              # Main package for REST API server
│   │   └── main.go
│   └── worker/           # Main package for worker service
│       └── main.go
│
├── internal/
│   ├── api/              # HTTP handlers and routing
│   │   ├── handler.go
│   │   └── router.go
│   ├── job/              # Job structs, Redis operations
│   │   ├── model.go
│   │   ├── queue.go
│   │   └── processor.go
│   ├── worker/           # Worker logic, retry, backoff
│   │   ├── worker.go
│   │   └── retry.go
│   ├── resilience/       # Circuit breaker, timeout wrappers
│   │   ├── breaker.go
│   │   └── timeout.go
│   └── observability/    # Logging, metrics, tracing
│       ├── logger.go
│       ├── metrics.go
│       └── tracing.go
│
├── configs/
│   ├── config.yaml       # App configs
│   └── k8s/              # Kubernetes manifests
│       ├── api-deployment.yaml
│       ├── worker-deployment.yaml
│       └── redis-deployment.yaml
│
├── docker/
│   ├── Dockerfile.api
│   ├── Dockerfile.worker
│   └── docker-compose.yaml
│
├── scripts/              # Dev/test scripts (e.g., DB seeding, cleanup)
│   └── seed_jobs.sh
│
├── test/
│   └── integration/      # Integration tests for job handling
│       ├── api_test.go
│       └── worker_test.go
│
├── go.mod
├── go.sum
└── README.md
