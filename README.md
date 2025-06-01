# go-queue


## Environment Variables

### API Service (`cmd/api/main.go`)
- `REDIS_ADDR` — Redis server address (e.g., `redis:6379`)
- `API_PORT` — Port for the API server (default: `8080`)

### Worker Service (`cmd/worker/main.go`)
- `REDIS_ADDR` — Redis server address (e.g., `redis:6379`)