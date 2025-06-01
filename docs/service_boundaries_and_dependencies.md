# Service Boundaries and Dependencies

## API Service

- **Role:** REST server for job submission.
- **Entrypoint:** `cmd/api/main.go`
- **Main Packages:**
    - `internal/api` (HTTP handlers, routing)
    - `internal/job` (shared job queue logic)
- **Environment Variables:**
    - `API_PORT` (default: 8080)
    - `REDIS_ADDR` (Redis server address)
- **Configuration:**
    - Redis address and API port are read from environment variables.

## Worker Service

- **Role:** Consumes jobs from Redis, processes with retry/backoff.
- **Entrypoint:** `cmd/worker/main.go`
- **Main Packages:**
    - `internal/worker` (worker logic)
    - `internal/job` (shared job queue logic)
- **Environment Variables:**
    - `REDIS_ADDR` (Redis server address)
- **Configuration:**
    - Redis address is read from environment variables.

## Shared Dependencies

- **Go Modules:**
    - Managed via `go.mod`.
- **Shared Code:**
    - `internal/job` (job structs, Redis queue logic)
- **Configuration:**
    - All services use environment variables for configuration (no hardcoded values).