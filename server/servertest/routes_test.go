package servertest

import (
	"io"
	"net/http"
	"time"

	"github.com/wspowell/log"
	"github.com/wspowell/spiderweb/endpoint"
	"github.com/wspowell/spiderweb/server"
)

func routes() *server.Server {
	serverConfig := &server.Config{
		LogConfig:    log.NewConfig(log.LevelDebug),
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	sample := server.New(serverConfig)

	sampleRoutes(sample)

	return sample
}

func sampleRoutes(sample *server.Server) {
	config := &endpoint.Config{
		LogConfig: &noopLogConfig{
			Config: log.NewConfig(log.LevelDebug),
		},
		Resources: map[string]interface{}{
			"datastore": &database{},
		},
		Timeout: 30 * time.Second,
	}

	sample.HandleNotFound(config, &noRoute{})
	sample.Handle(config, http.MethodPost, "/sample", &create{})
	sample.Handle(config, http.MethodGet, "/sample/{id}", &get{})
}

type noRoute struct{}

func (self *noRoute) Handle(ctx *endpoint.Context) (int, error) {
	return http.StatusNotFound, nil
}

type noopLogConfig struct {
	*log.Config
}

func (self *noopLogConfig) Out() io.Writer {
	return io.Discard
}
