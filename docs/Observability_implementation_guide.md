# Phase 6: Observability Implementation Guide

This guide provides step-by-step instructions for implementing observability in your Go/Redis queue system, focusing on demonstrating resilience, scalability, and high availability while keeping development efficient for a quick MVP.

## Step 1: Set Up Structured Logging

1. Add the Zap logger package to your project dependencies
2. Create a logger package with:
    - A global logger variable
    - An initialization function that creates a production logger
    - A getter function to access the logger

This provides a centralized, structured logging approach that can be used throughout your application.

## Step 2: Implement Metrics with Prometheus

1. Add Prometheus client libraries to your project dependencies
2. Create a metrics package with key metrics:
    - Counter for processed jobs (with status and job type labels)
    - Histogram for job processing time (with job type label)
    - Gauge for current queue length

These metrics will give you visibility into system performance and throughput.

## Step 3: Expose Metrics HTTP Endpoint

1. Add a `/metrics` endpoint to your HTTP server using the Prometheus handler
2. This makes your metrics available for scraping by Prometheus

## Step 4: Add OpenTelemetry Tracing

1. Add OpenTelemetry packages to your project dependencies
2. Create a tracing package with:
    - An initialization function that configures a Jaeger exporter
    - A trace provider setup with appropriate service name
    - A cleanup function for proper shutdown
    - A getter function to obtain a tracer

This establishes the foundation for distributed tracing across your system.

## Step 5: Implement HTTP Middleware for Tracing

1. Create a tracing middleware that:
    - Extracts any existing trace context from request headers
    - Creates a new span for each HTTP request
    - Passes the trace context to the next handler

This automatically traces all incoming HTTP requests with minimal code changes.

## Step 6: Implement Context Propagation Through Redis

1. Enhance your Job struct to include a tracing context field
2. Add methods to:
    - Extract tracing information from a context and store it with the job
    - Restore tracing context from a job when it's processed

This ensures trace context flows through your Redis queue between services.

## Step 7: Instrument API and Worker Components

1. For the API:
    - Create spans for key operations (job parsing, validation, enqueuing)
    - Include trace IDs in logs for correlation
    - Record metrics for job submission

2. For the Worker:
    - Create spans for job processing operations
    - Track processing time with metrics
    - Include trace IDs in all logs
    - Record job status in metrics

This provides end-to-end visibility of job lifecycle.

## Step 8: Add Kubernetes Manifests for Observability

1. Create Kubernetes manifests for:
    - Prometheus ConfigMap with scraping configuration
    - Prometheus Deployment with appropriate container settings
    - Prometheus Service to make it accessible

This sets up the infrastructure to collect and view your metrics.

## Step 9: Add Pod Annotations for Prometheus Scraping

1. Update your API and Worker pod definitions with annotations:
    - Enable Prometheus scraping
    - Specify metrics port and path

This allows Prometheus to automatically discover and scrape your services.

## Step 10: Test and Validate Observability

1. Submit test jobs through your API
2. Verify traces appear in Jaeger UI
3. Check metrics in Prometheus
4. Validate logs contain trace IDs for correlation
5. Test a failure scenario and verify it's properly observed

This confirms your observability implementation works end-to-end.

By following these steps, you'll implement a comprehensive observability solution that demonstrates your system's resilience, scalability, and high availability while keeping development efficient for your MVP.