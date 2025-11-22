SHELL := /bin/sh
BIN := go-metric-lab

.PHONY: run test vet lint fmt profile trace clean help

run: ## Start the HTTP server on :8080 (profiling on :6060)
	go run .

test: ## Run all unit tests with coverage output
	go test -cover ./...

vet: ## Run go vet for static analysis
	go vet ./...

lint: fmt vet ## Convenience target: gofmt + go vet

fmt: ## Format Go source files
	gofmt -w $$(go list -f '{{.Dir}}' ./...)

profile: ## Capture CPU profile via go tool pprof
	go tool pprof http://localhost:6060/debug/pprof/profile

trace: ## Record a 5-second trace and launch the trace viewer
	curl http://localhost:6060/debug/pprof/trace?seconds=5 -o trace.out
	go tool trace trace.out

clean: ## Remove generated artifacts
	$(RM) trace.out

help: ## Show available make targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-12s\033[0m %s\n", $$1, $$2}'

