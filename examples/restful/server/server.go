package server

import (
	"time"

	"github.com/wspowell/log"

	"github.com/wspowell/spiderweb/examples/restful/api"
	"github.com/wspowell/spiderweb/server/restful"
)

func New() *restful.Server {
	serverConfig := &restful.ServerConfig{
		LogConfig:    log.NewConfig(log.LevelDebug),
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	custom := restful.NewServer(serverConfig)

	api.Routes(custom)

	return custom
}
