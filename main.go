package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-metric-lab/diagnostics"
	"go-metric-lab/handlers"
	"go-metric-lab/metrics"
)

// main wires up the full lab so that observers can practice profiling, tracing,
// and metric-driven debugging on a realistic yet deterministic workload.
func main() {
	diagnostics.StartProfilingServer()

	mux := http.NewServeMux()

	mux.Handle("/api/work", metrics.InstrumentHandler(http.HandlerFunc(handlers.Work)))
	mux.Handle("/metrics", metrics.Handler())
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Println("go-metric-lab: HTTP server listening on :8080")
	log.Println("go-metric-lab: profiling via `go tool pprof http://localhost:6060/debug/pprof/profile`")
	log.Println("go-metric-lab: tracing via `curl http://localhost:6060/debug/pprof/trace?seconds=5 > trace.out && go tool trace trace.out`")
	log.Println("go-metric-lab: validation via `go test -cover ./...` and `go vet ./...` to keep the lab production-ready")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("go-metric-lab: http server failure: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("go-metric-lab: graceful shutdown failed: %v", err)
		os.Exit(1)
	}

	log.Println("go-metric-lab: shutdown complete")
}
