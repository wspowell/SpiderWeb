CURRENT_DIR = $(shell pwd)
FUZZ=false

build:
	go build ./...

test: build 
	GOBIN=$(CURRENT_DIR)/bin go get gotest.tools/gotestsum
	-FUZZ=$(FUZZ) $(CURRENT_DIR)/bin/gotestsum --format dots -- -count=1 -parallel 8 -race -v ./...

fuzz: FUZZ=true
fuzz: test