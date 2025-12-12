.PHONY: all test lint security build clean ci help

# Default Go version (matches CI)
GO_VERSION := 1.21

help:
	@echo "Available targets:"
	@echo "  make test      - Run all tests (v1, v2, demo)"
	@echo "  make lint      - Run golangci-lint (root, v2, demo)"
	@echo "  make security  - Run gosec security scanner"
	@echo "  make build     - Build all packages"
	@echo "  make ci        - Run full CI pipeline locally"
	@echo "  make clean     - Clean build artifacts"
	@echo ""
	@echo "Individual targets:"
	@echo "  make test-v1   - Run tests for v1 only"
	@echo "  make test-v2   - Run tests for v2 only"
	@echo "  make test-demo - Run tests for demo only"
	@echo "  make lint-v1   - Run lint for root/v1 only"
	@echo "  make lint-v2   - Run lint for v2 only"
	@echo "  make lint-demo - Run lint for demo only"

# Full CI pipeline
ci: lint test build
	@echo "CI pipeline completed successfully"

# All tests
test: test-v1 test-v2 test-demo
	@echo "All tests passed"

test-v1:
	@echo "Running tests (v1)..."
	go test -v -race ./...

test-v2:
	@echo "Running tests (v2)..."
	cd v2 && go test -v -race ./...

test-demo:
	@echo "Running tests (demo)..."
	cd demo && go test -v -race ./...

# All linting
lint: lint-v1 lint-v2 lint-demo
	@echo "All lint checks passed"

lint-v1:
	@echo "Running golangci-lint (root/v1)..."
	golangci-lint run --timeout=5m ./...

lint-v2:
	@echo "Running golangci-lint (v2)..."
	cd v2 && golangci-lint run --timeout=5m ./...

lint-demo:
	@echo "Running golangci-lint (demo)..."
	cd demo && golangci-lint run --timeout=5m ./...

# Security scan
security:
	@echo "Running gosec security scanner..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec -no-fail ./...; \
	else \
		echo "gosec not installed. Install with: go install github.com/securego/gosec/v2/cmd/gosec@latest"; \
		exit 1; \
	fi

# Build all packages
build: build-v1 build-v2 build-demo build-examples
	@echo "All packages built successfully"

build-v1:
	@echo "Building package (v1)..."
	go build -v ./...

build-v2:
	@echo "Building package (v2)..."
	cd v2 && go build -v ./...

build-demo:
	@echo "Building package (demo)..."
	cd demo && go build -v ./...

build-examples:
	@echo "Building examples..."
	cd examples && go build -v ./...

# Benchmarks
bench: bench-v1 bench-v2
	@echo "Benchmarks completed"

bench-v1:
	@echo "Running benchmarks (v1)..."
	go test -bench=. -benchmem -run=^$$ ./...

bench-v2:
	@echo "Running benchmarks (v2)..."
	cd v2 && go test -bench=. -benchmem -run=^$$ ./...

# Module verification
verify:
	@echo "Verifying modules..."
	go mod download
	go mod verify
	cd v2 && go mod download
	cd demo && go mod download

# Clean
clean:
	@echo "Cleaning..."
	go clean -cache -testcache
	rm -f coverage*.out
	rm -f v2/coverage*.out
	rm -f demo/coverage*.out
	rm -f benchmark*.txt
	rm -f v2/benchmark*.txt

# Coverage
coverage:
	@echo "Running tests with coverage..."
	go test -race -coverprofile=coverage-v1.out ./...
	cd v2 && go test -race -coverprofile=coverage-v2.out ./...
	cd demo && go test -race -coverprofile=coverage-demo.out ./...
	@echo "Coverage files generated: coverage-v1.out, v2/coverage-v2.out, demo/coverage-demo.out"

# Mod tidy check (like compatibility job in CI)
tidy-check:
	@echo "Checking go.mod files..."
	go mod tidy
	git diff --exit-code go.mod || (echo "go.mod needs updating" && exit 1)
	cd v2 && go mod tidy
	cd demo && go mod tidy
