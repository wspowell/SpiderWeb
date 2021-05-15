package test

import (
	"time"

	"github.com/wspowell/log"
	"github.com/wspowell/spiderweb/endpoint"
	"github.com/wspowell/spiderweb/http"
)

func Routes() *http.Server {
	serverConfig := &http.ServerConfig{
		LogConfig: &NoopLogConfig{
			Config: log.NewConfig(log.LevelFatal),
		},
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		EnablePprof:  false,
	}

	sample := http.NewServer(serverConfig)

	sampleRoutes(sample)

	return sample
}

func sampleRoutes(sample *http.Server) {
	config := &endpoint.Config{
		LogConfig: &NoopLogConfig{
			Config: log.NewConfig(log.LevelFatal),
		},
		Resources: map[string]interface{}{
			"datastore": &Database{},
		},
		Timeout: 30 * time.Second,
	}

	sample.HandleNotFound(config, &noRoute{})
	sample.Handle(config, http.MethodPost, "/sample", &Create{})
	sample.Handle(config, http.MethodGet, "/sample/{id}", &Get{})
}
