package resources

import (
	"github.com/wspowell/log"

	"github.com/wspowell/spiderweb/endpoint"
	"github.com/wspowell/spiderweb/server/restful"
	"github.com/wspowell/spiderweb/server/route"
)

func Routes(custom *restful.Server, config *endpoint.Config) {
	getConfig := config
	getConfig.LogConfig = log.NewConfig().WithLevel(log.LevelDebug)

	custom.Handle(config, route.Post("/resources", &postResource{}))
	custom.Handle(getConfig, route.Get("/resources/{id}", &getResource{}))
}
