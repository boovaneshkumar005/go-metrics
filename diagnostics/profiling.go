package diagnostics

import (
	"log"
	"net/http"
	"net/http/pprof"
)

// StartProfilingServer launches a dedicated :6060 server that hosts the
// /debug/pprof suite plus trace capture. Keeping this separate from the main
// API avoids perturbing latency-critical code paths.
func StartProfilingServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	go func() {
		addr := ":6060"
		log.Printf("diagnostics: pprof and trace server listening on %s", addr)
		log.Printf("diagnostics: CPU profile via `go tool pprof http://localhost:6060/debug/pprof/profile`")
		log.Printf("diagnostics: trace via `curl http://localhost:6060/debug/pprof/trace?seconds=5 > trace.out && go tool trace trace.out`")
		if err := http.ListenAndServe(addr, mux); err != nil && err != http.ErrServerClosed {
			log.Fatalf("diagnostics: profiling server failed: %v", err)
		}
	}()
}
