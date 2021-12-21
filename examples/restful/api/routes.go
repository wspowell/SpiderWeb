package api

import (
	"time"

	"github.com/wspowell/context"
	"github.com/wspowell/log"

	"github.com/wspowell/spiderweb/endpoint"
	"github.com/wspowell/spiderweb/examples/restful/api/resources"
	"github.com/wspowell/spiderweb/examples/restful/middleware"
	"github.com/wspowell/spiderweb/examples/restful/resources/db"
	"github.com/wspowell/spiderweb/httpstatus"
	"github.com/wspowell/spiderweb/server/restful"
)

func Routes(custom *restful.Server) {
	endpointConfig := &endpoint.Config{
		ErrorHandler:      middleware.ErrorJsonWithCodeResponse{},
		LogConfig:         log.NewConfig().WithLevel(log.LevelDebug),
		MimeTypeHandlers:  endpoint.NewMimeTypeHandlers(),
		RequestValidator:  middleware.ValidateNoopRequest{},
		ResponseValidator: middleware.ValidateNoopResponse{},
		Resources: map[string]any{
			"datastore": db.NewDatabase(),
		},
		Timeout: 30 * time.Second,
	}

	custom.HandleNotFound(endpointConfig, &noRoute{})
	resources.Routes(custom, endpointConfig)
}

type noRoute struct{}

func (self *noRoute) Handle(ctx context.Context) (int, error) {
	return httpstatus.NotFound, nil
}
