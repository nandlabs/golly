default: all

all: clean vet lint test build

.PHONY: clean
clean:
	@go mod tidy

.PHONY: build
build:
	@go build -v ./...

.PHONY: vet
vet:
	@go vet ./...

.PHONY: lint
lint:
	@if command -v golangci-lint > /dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed, skipping lint"; \
	fi

.PHONY: test
test:
	@go test -v -race ./...

.PHONY: test-cover
test-cover:
	@go test -race -cover -covermode=atomic -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out

.PHONY: test-cover-html
test-cover-html: test-cover
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: examples
examples:
	@go build -v ./examples/...
