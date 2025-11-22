# go-metric-lab

`go-metric-lab` is a production-grade observability playground that combines Prometheus metrics, Go pprof, Go trace, and custom latency instrumentation to demonstrate how principal engineers approach performance debugging.

## Why This Lab Exists

- **Demonstrate clean architecture** where handlers depend on well-factored diagnostics and metrics packages.
- **Teach holistic observability** by correlating metrics, logs, traces, and profiles.
- **Provide reproducible workloads** (`/api/work`) that simulate real bottlenecks (DB latency, CPU loops, payload building).

## Key Capabilities

- `/api/work` – orchestrates DB sleeps, CPU loops, and JSON encoding while tracking individual phases.
- `/metrics` – Prometheus endpoint with counters, histograms, and custom duration buckets.
- `/debug/pprof/*` and `/debug/pprof/trace` – profiling endpoints served on port 6060.
- `metrics.Track` – defer-friendly timer for structured logging + histograms.
- `metrics.LogSnapshot` – quick CLI inspection without Prometheus.

## Folder Structure

- `main.go` – wires HTTP server, middleware, graceful shutdown.
- `handlers/work.go` – synthetic workload with rich telemetry.
- `metrics/*` – Prometheus registry, timers, in-memory counters.
- `diagnostics/profiling.go` – dedicated profiling server.

## Running the Lab

```bash
go run .
# or use the makefile shortcuts
make run
```

## Observability Workflow

1. **Hit `/api/work`** repeatedly (e.g., `watch -n1 curl -s localhost:8080/api/work | jq`).
2. **Scrape `/metrics`** with Prometheus:
   - `rate(http_requests_total[1m])` vs `histogram_quantile` for latency.
   - `section_duration_seconds_bucket{section="db_query"}` reveals DB simulation spikes.
3. **Capture CPU profiles**:
   ```bash
   go tool pprof http://localhost:6060/debug/pprof/profile
   # use "top" for hotspots, "web" for flamegraph-like visualization.
   ```
4. **Record traces**:
   ```bash
   curl http://localhost:6060/debug/pprof/trace?seconds=5 -o trace.out
   go tool trace trace.out
   ```
5. **Debug loops**:
   - Inspect `work_loop_iterations_total` to correlate CPU loops with latency.
   - Compare `loop_checksum` between responses to detect logic changes.
6. **Analyze histograms**:
   ```promql
   histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le))
   histogram_quantile(0.99, sum(rate(section_duration_seconds_bucket{section="db_query"}[5m])) by (le))
   ```

## Performance Investigation Guidelines

- **Flamegraphs**: After `go tool pprof -http=:0 ...`, inspect wide frames; focus on `handlers.simulateLoop` for CPU issues.
- **Slow DB Detection**: Alert on `section_duration_seconds{section="db_query"}` P95; tie spikes back to upstream services.
- **Loop Debugging**: Use checksum + loop histogram to differentiate correctness from performance regressions.
- **Histogram Interpretation**: Follow bucket trends (0.1s, 0.25s, etc.) to see if tail latency worsens; combine with request volume.

## Quality Gates

```bash
make test    # go test ./...
make vet     # go vet ./...
make lint    # gofmt + static checks placeholder
```

Use these before pushing changes to ensure the lab stays production-ready.

