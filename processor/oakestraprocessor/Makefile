# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

# Project parameters
PKG_PATH=./internal
BENCH_TIME=1s
BENCH_COUNT=5

.PHONY: all test bench bench-filter bench-calculation test-verbose clean proto

# Default target
all: test

# Generate proto files
proto:
	protoc --go_out=. --go-grpc_out=. --go-grpc_opt=paths=source_relative --go_opt=paths=source_relative proto/*.proto

# Run all tests
test:
	$(GOTEST) $(PKG_PATH)

# Run tests with verbose output
test-verbose:
	$(GOTEST) -v $(PKG_PATH)

# Run all benchmarks
bench: bench-filter bench-calculation

# Run filter-specific benchmarks
bench-filter:
	$(GOTEST) -bench='^BenchmarkFilter|^BenchmarkMetricFilterStruct|^BenchmarkFilterScenarios' \
		-benchtime=$(BENCH_TIME) \
		-count=$(BENCH_COUNT) \
		-benchmem $(PKG_PATH)

# Run calculation-specific benchmarks
bench-calculation:
	$(GOTEST) -bench='^BenchmarkRegisterService' \
		-benchtime=$(BENCH_TIME) \
		-count=$(BENCH_COUNT) \
		-benchmem $(PKG_PATH)

# Clean test cache
clean:
	$(GOCMD) clean -testcache