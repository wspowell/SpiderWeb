package resources

import (
	"github.com/wspowell/log"
	"github.com/wspowell/spiderweb/endpoint"
	"github.com/wspowell/spiderweb/http"
	"github.com/wspowell/spiderweb/server/restful"
)

func Routes(custom *restful.Server, config *endpoint.Config) {
	getConfig := config
	getConfig.LogConfig = log.NewConfig(log.LevelDebug)

	custom.Handle(config, http.MethodPost, "/resources", &postResource{})
	custom.Handle(getConfig, http.MethodGet, "/resources/{id}", &getResource{})
}
