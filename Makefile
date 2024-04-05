default: all

all: clean test build

.PHONY: clean
clean:
	@go mod tidy

.PHONY: build
build:
	@go build -v ./...

.PHONY: test
test:
	@go test -v ./...

.PHONY: test-cover
test-cover:
	@go test -cover -covermode=atomic ./...
