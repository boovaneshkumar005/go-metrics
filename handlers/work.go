package handlers

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"

	"go-metric-lab/metrics"
)

const (
	minDBLatency = 50 * time.Millisecond
	maxDBLatency = 200 * time.Millisecond
	loopIters    = 100_000
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// WorkResponse returns all of the measured timings so practitioners can
// correlate user-visible latency with Prometheus histograms, flamegraphs,
// and trace captures.
type WorkResponse struct {
	Message    string             `json:"message"`
	TimingsMS  map[string]float64 `json:"timings_ms"`
	Iterations int                `json:"loop_iterations"`
	Checksum   uint64             `json:"loop_checksum"`
	Tips       map[string]string  `json:"observability_tips"`
	Timestamp  time.Time          `json:"timestamp"`
}

// Work orchestrates the simulated workload while emitting metrics.Track calls
// so each phase is independently measurable.
func Work(w http.ResponseWriter, r *http.Request) {
	totalStop := metrics.Track("work_total")
	defer totalStop()

	start := time.Now()
	timings := make(map[string]float64)

	dbDuration := simulateDB()
	timings["db_query_ms"] = toMillis(dbDuration)

	loopDuration, checksum := simulateLoop(loopIters)
	timings["loop_ms"] = toMillis(loopDuration)
	metrics.AddLoopIterations(loopIters)

	resp := WorkResponse{
		Message:    "Synthetic workload completed; inspect Prometheus, pprof, and trace outputs for deeper insights.",
		TimingsMS:  timings,
		Iterations: loopIters,
		Checksum:   checksum,
		Tips: map[string]string{
			"flamegraph": "Capture CPU profile via `go tool pprof -http=:0 http://localhost:6060/debug/pprof/profile`; wide frames reveal hot spots such as handlers.simulateLoop.",
			"slow_db":    "Alert on `section_duration_seconds{section=\"db_query\"}` P95; spikes signal real DB regressions.",
			"loop_debug": "Compare checksum between runs; divergence implies logic bugs while histogram shifts imply CPU regressions.",
			"histogram":  "Query `histogram_quantile(0.99, sum(rate(http_request_duration_seconds_bucket[5m])) by (le))` to catch tail latency.",
		},
		Timestamp: time.Now().UTC(),
	}

	timings["total_ms"] = toMillis(time.Since(start))

	payload, err := encodeResponse(&resp)
	if err != nil {
		metrics.CountJSONError()
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(payload); err != nil {
		log.Printf("handlers.work: failed to write response: %v", err)
	}
	metrics.RecordBusinessLatency("work_handler", time.Since(start))
}

func simulateDB() time.Duration {
	stop := metrics.Track("db_query")
	defer stop()

	delay := minDBLatency + time.Duration(rand.Int63n(int64(maxDBLatency-minDBLatency)))
	time.Sleep(delay)
	return delay
}

func simulateLoop(iterations int) (time.Duration, uint64) {
	stop := metrics.Track("loop")
	defer stop()

	start := time.Now()
	var checksum uint64
	for i := 0; i < iterations; i++ {
		value := uint64(i*i) ^ uint64(i>>2)
		checksum += value
	}
	return time.Since(start), checksum
}

func encodeResponse(resp *WorkResponse) ([]byte, error) {
	stop := metrics.Track("json_encode")
	defer stop()

	return json.Marshal(resp)
}

func toMillis(d time.Duration) float64 {
	return float64(d.Microseconds()) / 1000.0
}
