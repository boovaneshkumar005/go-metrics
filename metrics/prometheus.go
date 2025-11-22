package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	registry = prometheus.NewRegistry()

	requestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests, partitioned by method and path.",
		},
		[]string{"method", "path"},
	)

	errorCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_errors_total",
			Help: "Total number of HTTP errors by method, path, and status.",
		},
		[]string{"method", "path", "status"},
	)

	requestLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Observed HTTP request duration.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	sectionLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "section_duration_seconds",
			Help:    "Latency for custom code sections recorded via metrics.Track.",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2},
		},
		[]string{"section"},
	)

	businessLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "business_latency_seconds",
			Help:    "High-level latency per logical business operation.",
			Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2, 5},
		},
		[]string{"operation"},
	)

	loopIterations = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "work_loop_iterations_total",
			Help: "Total simulated CPU loop iterations across all requests.",
		},
	)

	jsonErrorCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "json_encode_errors_total",
			Help: "Number of JSON encoding failures inside handlers.",
		},
	)
)

func init() {
	registry.MustRegister(requestCounter, errorCounter, requestLatency, sectionLatency, businessLatency, loopIterations, jsonErrorCounter)
}

// Handler exposes /metrics for Prometheus scrapers.
func Handler() http.Handler {
	return promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
}

// InstrumentHandler instruments every HTTP request, recording counters and
// latency histograms along with error rates for status >= 500.
func InstrumentHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()

		next.ServeHTTP(recorder, r)

		duration := time.Since(start)
		labels := []string{r.Method, r.URL.Path}
		requestCounter.WithLabelValues(labels...).Inc()
		requestLatency.WithLabelValues(labels...).Observe(duration.Seconds())

		if recorder.status >= http.StatusInternalServerError {
			errorCounter.WithLabelValues(append(labels, strconv.Itoa(recorder.status))...).Inc()
		}
	})
}

// ObserveSectionLatency feeds the section histogram so engineers can visualize
// each tracked code block. Example dashboards:
//   * Flamegraph context: graph section_duration_seconds_bucket filtered by section="loop"
//   * Slow DB detection: alert on P95 of section="db_query"
func ObserveSectionLatency(section string, duration time.Duration) {
	sectionLatency.WithLabelValues(section).Observe(duration.Seconds())
}

// ObserveBusinessLatency records high-level handler timings for
// histogram_quantile queries.
func ObserveBusinessLatency(operation string, duration time.Duration) {
	businessLatency.WithLabelValues(operation).Observe(duration.Seconds())
}

// ObserveLoopIterations exposes the amount of simulated CPU work as a custom
// metric so dashboards can correlate CPU usage with request load.
func ObserveLoopIterations(iterations float64) {
	loopIterations.Add(iterations)
}

// IncJSONError increments the JSON-specific counter so failed encodes are
// visible alongside HTTP failures.
func IncJSONError() {
	jsonErrorCounter.Inc()
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}
