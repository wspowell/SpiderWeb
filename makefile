CURRENT_DIR = $(shell pwd)
FUZZ=false

build:
	go build ./...

test: build 
	GOBIN=$(CURRENT_DIR)/bin go get gotest.tools/gotestsum
	-FUZZ=$(FUZZ) $(CURRENT_DIR)/bin/gotestsum --format dots -- -count=1 -parallel 8 -race -v ./...
	-FUZZ=$(FUZZ) $(CURRENT_DIR)/bin/gotestsum --format dots -- -count=1 -parallel 8 -race -v -tags release ./...

	# Ensure benchmarks have no race conditions
	go test -bench=. -benchmem -count=1 -parallel 8 -cpu 8 -race ./...
	go test -bench=. -benchmem -count=1 -parallel 8 -cpu 8 -race -tags release ./...

fuzz: FUZZ=true
fuzz: test

bench: build
	go test -bench=. -benchmem -count=1 -parallel 8 -cpu 8 ./...
	go test -bench=. -benchmem -count=1 -parallel 8 -cpu 8 -tags release ./...

bench-cpu:
	go test -tags release -c ./server/servertest/
	./servertest.test -test.benchtime=20s -test.benchmem -test.count=1 -test.parallel 8 -test.cpu 8  -test.bench "(Benchmark_SpiderWeb_POST_latency)"

profile-cpu:
	go tool pprof -seconds 10 -png ./servertest.test http://localhost:6060/debug/pprof/profile > profile-cpu.png