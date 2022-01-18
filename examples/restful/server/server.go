package server

import (
	"time"

	"github.com/wspowell/log"

	"github.com/wspowell/spiderweb/examples/restful/api"
	"github.com/wspowell/spiderweb/examples/restful/resources"
	"github.com/wspowell/spiderweb/server/restful"
)

func New(datastore resources.Datastore) *restful.Server {
	serverConfig := &restful.ServerConfig{
		LogConfig:    log.NewConfig().WithLevel(log.LevelDebug),
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	custom := restful.NewServer(serverConfig)

	api.Routes(custom, datastore)

	return custom
}
