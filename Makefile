.PHONY: install-tools lint test test-verbose format benchmark

install-tools:
	# Install linting tools
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.49.0
	go install mvdan.cc/gofumpt@latest
	go install github.com/segmentio/golines@latest

lint:
	golangci-lint run -c ./golangci.yml ./...

format:
	gofumpt -l -w -extra .
	golines . -w

test:
	go test -race -test.timeout 120s -count=1 ./...

test-verbose:
	go test -race -test.timeout 120s -v -cover -count=1 ./...

benchmark:
	go test -bench=. -benchmem ./...
