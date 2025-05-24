# Prompt for Writing Go Code for the Resilient Job Queue & Worker System

## Overview

You are to write Go code for a project with these characteristics:

- **Project:** Fault-tolerant, scalable job queue and worker system using Go and Redis.
- **Features:** REST API for job submission, Redis-backed job queue, worker service with concurrency, retry, backoff, graceful shutdown, resilience patterns (timeout, circuit breaker, dead-letter queue), containerization, Kubernetes deployment, and full observability (metrics, logs, tracing).

## Development Plan

- The project is developed in phases (see the development plan for details).
- Each phase has clear goals and tasks (e.g., REST API, worker, resilience, containerization, Kubernetes, observability).
- Only implement the current phase; do not anticipate future phases unless specified.

## Developer's Guide

- Write idiomatic, high-quality, and consistent Go code.
- Use the standard library and idiomatic Go patterns.
- Use embedding where efficient to promote code reuse and reduce boilerplate.
- Use the testify assertion library for unit tests.
- Use table-driven tests for multiple scenarios.
- Mock only true external dependencies (e.g., Redis, network calls, processors).
- Cover both success and failure paths, including edge cases.
- Avoid global state and ensure tests are independent.
- Use `t.Parallel()` where safe.
- Add comments for complex logic.
- Include tests for concurrency aspects where relevant.
- Do not include integration tests or benchmarks.
- Use clean architecture: separate handler, service, and repository layers.
- Use interfaces, dependency injection, and robust error handling.
- Ensure safe use of goroutines, channels, and mutexes.
- Profile and optimize code using benchmarking and tracing tools.
- Manage goroutines and channels effectively.
- Use worker pools, mutexes, and semaphores for concurrency safety.
- Ensure statelessness and proper connection pooling.
- Apply context timeouts for all external calls.
- Implement retries with exponential backoff.
- Handle graceful shutdown and robust error handling.
- Use structured logging (`zap`, `logrus`, or `zerolog`).
- Expose a Prometheus `/metrics` endpoint.
- Integrate basic OpenTelemetry tracing (optional).
- Use multi-stage Docker builds, health checks, and environment configuration.
- Provide Kubernetes manifests for deployment (optional).

## Instructions

1. Implement only the current phase as described in the development plan.
2. Follow the project overview and developer's guide strictly.
3. If any requirements are unclear or there are multiple valid approaches, ask for clarification before proceeding.
4. Write idiomatic, maintainable, and well-documented Go code.
5. Include high-quality unit tests as specified in the developer's guide.
6. Use actual struct definitions from the codebase in your tests.
7. Do not define or use custom test structs with similar fields.
8. Do not include integration tests or benchmarks.

---

**Example Usage:**

> Implement Phase 1: Foundation Setup.
> - Scaffold a Go module.
> - Create a simple REST API for job enqueue.
> - Connect to Redis using go-redis.
> - Implement a basic job enqueue endpoint.
> - Review concurrency fundamentals in Go.
>
> If anything is unclear (e.g., job struct fields, API endpoint details), ask for clarification before proceeding.
