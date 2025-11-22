package metrics

import (
	"log"
	"sync"
	"time"
)

var (
	mu             sync.RWMutex
	sectionTotals  = make(map[string]time.Duration)
	sectionSamples = make(map[string]int64)
	customCounters = make(map[string]int64)
)

func recordDuration(section string, d time.Duration) {
	mu.Lock()
	defer mu.Unlock()
	sectionTotals[section] += d
	sectionSamples[section]++
}

// RecordBusinessLatency allows handlers to push ad-hoc latencies when they want
// to correlate API timing with Prometheus histograms and logs.
func RecordBusinessLatency(operation string, d time.Duration) {
	recordDuration(operation, d)
	ObserveBusinessLatency(operation, d)
}

// LogSnapshot helps during demos when you want to dump in-memory counters to
// standard output for quick inspection without hitting Prometheus.
func LogSnapshot() {
	mu.RLock()
	defer mu.RUnlock()
	for section, total := range sectionTotals {
		count := sectionSamples[section]
		avg := time.Duration(0)
		if count > 0 {
			avg = time.Duration(int64(total) / count)
		}
		log.Printf("metrics.snapshot: section=%s count=%d avg_ms=%.2f", section, count, float64(avg.Microseconds())/1000.0)
	}
	for name, value := range customCounters {
		log.Printf("metrics.snapshot: counter=%s value=%d", name, value)
	}
}

// CountJSONError increments a simple in-memory counter and publishes the same
// detail to Prometheus so failed responses can be graphed.
func CountJSONError() {
	incrementCounter("json_encode_error")
	IncJSONError()
}

// AddLoopIterations lets business logic communicate how heavy CPU loops were.
func AddLoopIterations(iterations int) {
	incrementCounter("loop_iterations")
	ObserveLoopIterations(float64(iterations))
}

func incrementCounter(name string) {
	mu.Lock()
	defer mu.Unlock()
	customCounters[name]++
}
