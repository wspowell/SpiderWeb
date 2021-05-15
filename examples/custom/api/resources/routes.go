package resources

import (
	"github.com/wspowell/log"
	"github.com/wspowell/spiderweb/endpoint"
	"github.com/wspowell/spiderweb/http"
)

func Routes(custom *http.Server, config *endpoint.Config) {
	getConfig := config
	getConfig.LogConfig = log.NewConfig(log.LevelDebug)

	custom.Handle(config, http.MethodPost, "/resources", &postResource{})
	custom.Handle(getConfig, http.MethodGet, "/resources/{id}", &getResource{})
}
