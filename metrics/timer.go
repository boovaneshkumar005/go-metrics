package metrics

import (
	"log"
	"time"
)

// Track wraps a code block and logs how long it took. Pair it with defer so
// even early returns emit telemetry:
//   defer metrics.Track("db_query")()
// Track also feeds Prometheus histograms for latency-based dashboards.
func Track(section string) func() {
	start := time.Now()
	log.Printf("metrics.track: section=%s phase=start", section)
	return func() {
		duration := time.Since(start)
		log.Printf("metrics.track: section=%s duration_ms=%.3f", section, float64(duration.Microseconds())/1000.0)
		recordDuration(section, duration)
		ObserveSectionLatency(section, duration)
	}
}
