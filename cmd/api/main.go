package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sat-h/go-queue/internal/api"
	"github.com/sat-h/go-queue/internal/job"
	"github.com/sat-h/go-queue/internal/observability/logger"
	"github.com/sat-h/go-queue/internal/observability/metrics"
	"github.com/sat-h/go-queue/internal/observability/tracing"
	"go.uber.org/zap"
)

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func readinessHandler(w http.ResponseWriter, _ *http.Request) {
	// Optionally, add deeper checks (e.g., Redis connectivity)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ready"))
}

func main() {
	// Initialize observability components
	logger.Init()
	defer logger.Sync()

	metrics.Init()

	// Initialize tracing with service name
	if err := tracing.Init("go-queue-api"); err != nil {
		logger.Error("Failed to initialize tracing", zap.Error(err))
	}

	// Create context with cancellation for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Ensure tracing is properly shutdown
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*1000000000) // 5 seconds
		defer cancel()
		if err := tracing.Shutdown(shutdownCtx); err != nil {
			logger.Error("Error shutting down tracer", zap.Error(err))
		}
	}()

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
			RedisMode:  "sentinel",
			RedisAddrs: sentinels,
			MasterName: masterName,
			Password:   os.Getenv("REDIS_PASSWORD"),
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
	router := api.NewRouter(handler)

	mux := http.NewServeMux()
	mux.Handle("/healthz", http.HandlerFunc(healthHandler))
	mux.Handle("/readyz", http.HandlerFunc(readinessHandler))
	// Expose Prometheus metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/", router)

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info("API running", zap.String("port", port))
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// Start server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	<-ctx.Done()
	logger.Info("Shutting down API server...")

	// Create a shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*1000000000) // 10 seconds
	defer shutdownCancel()

	// Shutdown the server
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server shutdown error", zap.Error(err))
	}

	logger.Info("API server stopped")
}
