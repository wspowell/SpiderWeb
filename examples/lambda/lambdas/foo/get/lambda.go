package main

import (
	"time"

	"github.com/wspowell/logging"
	"github.com/wspowell/spiderweb/endpoint"
	"github.com/wspowell/spiderweb/lambda"
)

func main() {
	config := &endpoint.Config{
		LogConfig: logging.NewConfig(logging.LevelDebug),
		Timeout:   30 * time.Second,
	}

	handler := lambda.New(config, &get{})
	handler.Start()
}
