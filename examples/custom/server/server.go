package server

import (
	"time"

	"github.com/wspowell/logging"
	"github.com/wspowell/spiderweb/server"
)

func New() *server.Server {
	serverConfig := &server.ServerConfig{
		LogConfig:    logging.NewConfig(logging.LevelDebug, map[string]interface{}{}),
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	server := server.NewServer(serverConfig)

	return server
}
