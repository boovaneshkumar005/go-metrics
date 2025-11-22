# Example Investigation Flow

This walkthrough shows how to use `go-metric-lab` to debug a synthetic latency spike end-to-end.

## 1. Start the Lab

```bash
make run
```

Profiling endpoints are now at `http://localhost:6060/debug/pprof/`.

## 2. Generate Load

```bash
watch -n0.5 'curl -s http://localhost:8080/api/work | jq .timings_ms'
```

Look for slow sections:
- `db_query_ms` rising ⇒ simulated DB issues.
- `loop_ms` rising ⇒ CPU loop hot path.

## 3. Inspect Prometheus Metrics

Scrape `/metrics` (e.g., via `curl localhost:8080/metrics | grep section_duration_seconds`).

PromQL snippets:

```promql
rate(http_requests_total[1m])
histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le))
histogram_quantile(0.99, sum(rate(section_duration_seconds_bucket{section="db_query"}[5m])) by (le))
```

## 4. Capture CPU Profile

```bash
go tool pprof -http=:0 http://localhost:6060/debug/pprof/profile
```

- Wide stack frames around `handlers.simulateLoop` imply CPU pressure.
- Compare before/after optimizations to validate improvements.

## 5. Record Execution Trace

```bash
make trace
```

Use the trace viewer to:
- Spot scheduling delays.
- Validate goroutine behavior during DB sleeps.

## 6. Troubleshoot JSON Encoding

- Check `json_encode_errors_total` for failures.
- Inspect `metrics.track: section=json_encode` logs for high latency.

## 7. Snapshot In-Memory Stats

Insert `metrics.LogSnapshot()` in a debug path to dump averages:

```
metrics.snapshot: section=db_query count=120 avg_ms=0.08
metrics.snapshot: counter=loop_iterations value=12000000
```

## 8. Interpret Results

- If DB histogram p95 > 200ms, investigate upstream storage or network.
- If loop histogram drifts upward, optimize algorithm or reduce iterations.
- Correlate Prometheus data with pprof to confirm root cause before shipping fixes.

