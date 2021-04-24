package main

import (
	"time"

	"github.com/wspowell/log"
	"github.com/wspowell/spiderweb/endpoint"
	"github.com/wspowell/spiderweb/lambda"
)

func main() {
	config := &endpoint.Config{
		LogConfig: log.NewConfig(log.LevelDebug),
		Timeout:   30 * time.Second,
	}

	handler := lambda.New(config, "/foo", &get{})
	handler.Start()
}
