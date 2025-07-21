package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// jobsProcessed counts the number of processed jobs with status and job type labels
	jobsProcessed *prometheus.CounterVec

	// jobProcessingTime measures the time taken to process jobs with job type label
	jobProcessingTime *prometheus.HistogramVec

	// queueLength tracks the number of jobs in the queue
	queueLength prometheus.Gauge

	// redisConnectionFailures counts the number of Redis connection failures
	redisConnectionFailures *prometheus.CounterVec

	// redisReconnectionAttempts counts the number of Redis reconnection attempts
	redisReconnectionAttempts prometheus.Counter

	// redisReconnectionSuccess counts the number of successful Redis reconnections
	redisReconnectionSuccess prometheus.Counter

	// jobDeduplicationEvents counts the number of prevented duplicate job processing events
	jobDeduplicationEvents *prometheus.CounterVec

	// workerRecoveryTime measures the time taken for a worker to recover after failures
	workerRecoveryTime *prometheus.HistogramVec

	once sync.Once
)

// Init initializes the metrics collection system
func Init() {
	once.Do(func() {
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
				Name:    "worker_recovery_time_seconds",
				Help:    "Time taken for worker to recover after failures",
				Buckets: prometheus.ExponentialBuckets(0.1, 2, 10), // 0.1s to ~100s
			},
			[]string{"failure_type"}, // "redis", "processing", etc.
		)
	})
}

// JobProcessed increments the jobs processed counter
func JobProcessed(status, jobType string) {
	if jobsProcessed != nil {
		jobsProcessed.WithLabelValues(status, jobType).Inc()
	}
}

// ObserveJobProcessingTime records the time taken to process a job
func ObserveJobProcessingTime(jobType string, durationSeconds float64) {
	if jobProcessingTime != nil {
		jobProcessingTime.WithLabelValues(jobType).Observe(durationSeconds)
	}
}

// SetQueueLength updates the queue length gauge
func SetQueueLength(length float64) {
	if queueLength != nil {
		queueLength.Set(length)
	}
}

// RecordRedisConnectionFailure records a Redis connection failure
func RecordRedisConnectionFailure(connectionType string) {
	if redisConnectionFailures != nil {
		redisConnectionFailures.WithLabelValues(connectionType).Inc()
	}
}

// RecordRedisReconnectionAttempt records a Redis reconnection attempt
func RecordRedisReconnectionAttempt() {
	if redisReconnectionAttempts != nil {
		redisReconnectionAttempts.Inc()
	}
}

// RecordRedisReconnectionSuccess records a successful Redis reconnection
func RecordRedisReconnectionSuccess() {
	if redisReconnectionSuccess != nil {
		redisReconnectionSuccess.Inc()
	}
}

// RecordJobDeduplication records a job deduplication event
func RecordJobDeduplication(jobType string) {
	if jobDeduplicationEvents != nil {
		jobDeduplicationEvents.WithLabelValues(jobType).Inc()
	}
}

// ObserveWorkerRecoveryTime records the time taken for a worker to recover
func ObserveWorkerRecoveryTime(failureType string, durationSeconds float64) {
	if workerRecoveryTime != nil {
		workerRecoveryTime.WithLabelValues(failureType).Observe(durationSeconds)
	}
}

// GetJobsProcessedCounter returns the jobsProcessed counter for testing
func GetJobsProcessedCounter() *prometheus.CounterVec {
	return jobsProcessed
}

// GetJobProcessingTimeHistogram returns the jobProcessingTime histogram for testing
func GetJobProcessingTimeHistogram() *prometheus.HistogramVec {
	return jobProcessingTime
}

// GetQueueLengthGauge returns the queueLength gauge for testing
func GetQueueLengthGauge() prometheus.Gauge {
	return queueLength
}

// GetRedisConnectionFailuresCounter returns the redisConnectionFailures counter for testing
func GetRedisConnectionFailuresCounter() *prometheus.CounterVec {
	return redisConnectionFailures
}

// GetRedisReconnectionAttemptsCounter returns the redisReconnectionAttempts counter for testing
func GetRedisReconnectionAttemptsCounter() prometheus.Counter {
	return redisReconnectionAttempts
}

// GetRedisReconnectionSuccessCounter returns the redisReconnectionSuccess counter for testing
func GetRedisReconnectionSuccessCounter() prometheus.Counter {
	return redisReconnectionSuccess
}

// GetJobDeduplicationEventsCounter returns the jobDeduplicationEvents counter for testing
func GetJobDeduplicationEventsCounter() *prometheus.CounterVec {
	return jobDeduplicationEvents
}

// GetWorkerRecoveryTimeHistogram returns the workerRecoveryTime histogram for testing
func GetWorkerRecoveryTimeHistogram() *prometheus.HistogramVec {
	return workerRecoveryTime
}
