.PHONY: help install-tools lint test test-verbose format benchmark
.SILENT: help install-tools lint test test-verbose format benchmark

help:
	grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

install-tools: ## Install linting tools
	# Install linting tools
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.49.0
	go install mvdan.cc/gofumpt@latest
	go install github.com/segmentio/golines@latest

lint: ## Run golangci linter
	golangci-lint run -c ./golangci.yml ./...

format: ## Format code
	gofumpt -l -w -extra .
	golines . -w

test: ## Run tests
	go test -race -test.timeout 120s -count=1 ./...

test-verbose: ## Run tests with verbose output
	go test -race -test.timeout 120s -v -cover -count=1 ./...

benchmark: ## Run benchmarks
	go test -bench=. -benchmem ./...
