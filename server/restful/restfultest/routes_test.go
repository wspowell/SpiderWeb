package restfultest_test

import (
	"fmt"
	"time"

	"github.com/wspowell/log"

	"github.com/wspowell/spiderweb/examples/restful/resources"
	"github.com/wspowell/spiderweb/handler"
	"github.com/wspowell/spiderweb/httpmethod"
	"github.com/wspowell/spiderweb/server/restful"
	"github.com/wspowell/spiderweb/test"
)

func RoutesTest(database resources.Datastore) *restful.Server {
	fmt.Printf("%p\n", database)
	serverConfig := &restful.ServerConfig{
		LogConfig: &test.NoopLogConfig{
			Config: log.NewConfig().WithLevel(log.LevelDebug),
		},
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		EnablePprof:  false,
	}

	sample := restful.NewServer(serverConfig)

	sampleRoutes(sample, database)

	return sample
}

func sampleRoutes(sample *restful.Server, database resources.Datastore) {
	sample.HandleNotFound(handler.NewHandle(test.NoRoute{}))
	sample.Handle(httpmethod.Post, "/sample", handler.NewHandle(test.Create{}))
	sample.Handle(httpmethod.Get, "/sample/{id}", handler.NewHandle(test.Get{
		Db: database,
	}))
}
