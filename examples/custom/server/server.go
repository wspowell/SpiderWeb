package server

import (
	"time"

	"github.com/wspowell/log"
	"github.com/wspowell/spiderweb/examples/custom/api"
	"github.com/wspowell/spiderweb/http"
)

func New() *http.Server {
	serverConfig := &http.ServerConfig{
		LogConfig:    log.NewConfig(log.LevelDebug),
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	custom := http.NewServer(serverConfig)

	api.Routes(custom)

	return custom
}
