package restfultest_test

import (
	"time"

	"github.com/wspowell/log"
	"github.com/wspowell/spiderweb/endpoint"
	"github.com/wspowell/spiderweb/http"
	"github.com/wspowell/spiderweb/server/restful"
	"github.com/wspowell/spiderweb/test"
)

func Routes() *restful.Server {
	serverConfig := &restful.ServerConfig{
		LogConfig: &test.NoopLogConfig{
			Config: log.NewConfig(log.LevelFatal),
		},
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		EnablePprof:  false,
	}

	sample := restful.NewServer(serverConfig)

	sampleRoutes(sample)

	return sample
}

func sampleRoutes(sample *restful.Server) {
	config := &endpoint.Config{
		LogConfig: &test.NoopLogConfig{
			Config: log.NewConfig(log.LevelFatal),
		},
		Resources: map[string]interface{}{
			"datastore": &test.Database{},
		},
		Timeout: 30 * time.Second,
	}

	sample.HandleNotFound(config, &test.NoRoute{})
	sample.Handle(config, http.MethodPost, "/sample", &test.Create{})
	sample.Handle(config, http.MethodGet, "/sample/{id}", &test.Get{})
}
