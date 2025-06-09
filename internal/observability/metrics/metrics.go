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
