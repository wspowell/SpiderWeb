CURRENT_DIR = $(shell pwd)
FUZZ=false

build:
	go build ./...

test: build
	# Run with benchmarks to ensure they have no race conditions
	-FUZZ=$(FUZZ) $(CURRENT_DIR)/bin/gotestsum --format dots -- -bench=. -benchmem -count=1 -parallel 8 -race -v ./...
	-FUZZ=$(FUZZ) $(CURRENT_DIR)/bin/gotestsum --format dots -- -bench=. -benchmem -count=1 -parallel 8 -race -v -tags release ./...

fuzz: FUZZ=true
fuzz: test

bench: build
	go test -run== -bench=. -benchmem -count=3 -parallel 8 -cpu 8 ./...
	go test -run== -bench=. -benchmem -count=3 -parallel 8 -cpu 8 -tags release ./...

bench-latency:
	go test -tags release -c ./server/servertest/
	./servertest.test -test.benchtime=60s -test.benchmem -test.count=1 -test.parallel 8 -test.cpu 8 -test.cpuprofile cpu.prof -test.bench "(Benchmark_SpiderWeb_POST_latency)"
	go tool pprof -png ./cpu.prof > cpu_latency.png

bench-throughput:
	go test -tags release -c ./server/servertest/
	./servertest.test -test.benchtime=60s -test.benchmem -test.count=1 -test.parallel 8 -test.cpu 8 -test.cpuprofile cpu.prof -test.bench "(Benchmark_SpiderWeb_POST_throughput)"
	go tool pprof -png ./cpu.prof > cpu_throughput.png
