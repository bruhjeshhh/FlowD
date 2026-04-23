package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	JobsEnqueued = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "flowd_jobs_enqueued_total",
			Help: "Total number of jobs enqueued",
		},
		[]string{"type"},
	)

	JobsProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "flowd_jobs_processed_total",
			Help: "Total number of jobs processed",
		},
		[]string{"type", "status"},
	)

	JobDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "flowd_job_duration_seconds",
			Help:    "Job processing duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"type"},
	)

	JobsInQueue = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "flowd_jobs_in_queue",
			Help: "Number of jobs currently in queue by status",
		},
		[]string{"status"},
	)

	WorkersActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "flowd_workers_active",
			Help: "Number of currently active workers",
		},
	)

	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "flowd_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "flowd_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	JobRetries = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "flowd_job_retries_total",
			Help: "Total number of job retries",
		},
		[]string{"type"},
	)

	DeadLetterJobs = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "flowd_dead_letter_jobs_total",
			Help: "Total number of jobs moved to dead letter queue",
		},
		[]string{"type"},
	)
)
)
