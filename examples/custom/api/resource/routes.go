package resource

import (
	"net/http"

	"github.com/wspowell/spiderweb/endpoint"
	"github.com/wspowell/spiderweb/server"
)

func Routes(custom *server.Server, config *endpoint.Config) {
	custom.Handle(config, http.MethodPost, "/resources", &postResource{})
	custom.Handle(config, http.MethodGet, "/resources/{id}", &getResource{})
}
