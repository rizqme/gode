# Gode Makefile

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=gode
BINARY_PATH=./cmd/gode

# Build the CLI binary
build:
	$(GOBUILD) -o $(BINARY_NAME) $(BINARY_PATH)

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

# Run all tests
test:
	$(GOTEST) -v ./...

# Run unit tests only
test-unit:
	$(GOTEST) -v ./internal/... ./pkg/...

# Run integration tests only
test-integration:
	$(GOTEST) -v ./tests/integration/...

# Run e2e tests only
test-e2e:
	$(GOTEST) -v ./tests/e2e/...

# Run benchmarks
bench:
	$(GOTEST) -bench=. -benchmem ./...

# Run unit benchmarks only
bench-unit:
	$(GOTEST) -bench=. -benchmem ./internal/... ./pkg/...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Run specific test files
test-vm:
	$(GOTEST) -v ./internal/runtime/vm_test.go ./internal/runtime/vm.go ./internal/runtime/goja_vm.go

test-runtime:
	$(GOTEST) -v ./internal/runtime/runtime_test.go ./internal/runtime/runtime.go

test-modules:
	$(GOTEST) -v ./internal/modules/manager_test.go ./internal/modules/manager.go

test-config:
	$(GOTEST) -v ./pkg/config/package_test.go ./pkg/config/package.go

test-module-resolution:
	$(GOTEST) -v ./tests/integration/module_resolution_test.go

# Run specific test by name
test-name:
	@echo "Usage: make test-name TEST=TestName"
	$(GOTEST) -v -run $(TEST) ./...

# Format code
fmt:
	$(GOCMD) fmt ./...

# Vet code
vet:
	$(GOCMD) vet ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Install the binary
install: build
	cp $(BINARY_NAME) $(GOPATH)/bin/

# Run examples with the built binary
run-example:
	./$(BINARY_NAME) run examples/simple.js

# Quick test of the built binary
quick-test: build
	./$(BINARY_NAME) version
	./$(BINARY_NAME) run testdata/simple.js

# Full test suite (build, test, e2e)
test-all: build test test-e2e

# Check everything (format, vet, test, build)
check: fmt vet test build

# Development workflow
dev: fmt vet test build run-example

# CI workflow
ci: deps fmt vet test test-integration test-e2e build

.PHONY: build clean test test-unit test-integration test-e2e bench bench-unit test-coverage \
        test-vm test-runtime test-modules test-config test-module-resolution test-name fmt vet lint deps install \
        run-example quick-test test-all check dev ci