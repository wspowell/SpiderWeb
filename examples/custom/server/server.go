package server

import (
	"time"

	"github.com/wspowell/logging"
	"github.com/wspowell/spiderweb/examples/custom/api"
	"github.com/wspowell/spiderweb/server"
)

func New() *server.Server {
	serverConfig := &server.Config{
		LogConfig:    logging.NewConfig(logging.LevelDebug, map[string]interface{}{}),
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	custom := server.New(serverConfig)

	api.Routes(custom)

	return custom
}
