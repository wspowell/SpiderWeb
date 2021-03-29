CURRENT_DIR = $(shell pwd)
FUZZ=false

build:
	go build ./...

test: build 
	GOBIN=$(CURRENT_DIR)/bin go get gotest.tools/gotestsum
	-FUZZ=$(FUZZ) $(CURRENT_DIR)/bin/gotestsum --format dots -- -count=1 -parallel 8 -race -v ./...
	-FUZZ=$(FUZZ) $(CURRENT_DIR)/bin/gotestsum --format dots -- -count=1 -parallel 8 -race -v -tags release ./...

fuzz: FUZZ=true
fuzz: test

bench: build
	go test  -bench=. -benchmem -count=1 -parallel 8 -cpu 8 -race ./spiderwebtest
	go test  -bench=. -benchmem -count=1 -parallel 8 -cpu 8 -race -tags release ./spiderwebtest