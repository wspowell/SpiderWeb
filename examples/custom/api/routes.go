package api

import (
	"time"

	"github.com/wspowell/context"
	"github.com/wspowell/log"
	"github.com/wspowell/spiderweb/endpoint"
	"github.com/wspowell/spiderweb/examples/custom/api/resources"
	"github.com/wspowell/spiderweb/examples/custom/middleware"
	"github.com/wspowell/spiderweb/examples/custom/resources/db"
	"github.com/wspowell/spiderweb/http"
)

func Routes(custom *http.Server) {
	endpointConfig := &endpoint.Config{
		Auther:       middleware.AuthNoop{},
		ErrorHandler: middleware.ErrorJsonWithCodeResponse{},
		LogConfig: &middleware.NoopLogConfig{
			Config: log.NewConfig(log.LevelDebug),
		},
		MimeTypeHandlers:  endpoint.NewMimeTypeHandlers(),
		RequestValidator:  middleware.ValidateNoopRequest{},
		ResponseValidator: middleware.ValidateNoopResponse{},
		Resources: map[string]interface{}{
			"datastore": db.NewDatabase(),
		},
		Timeout: 30 * time.Second,
	}

	custom.HandleNotFound(endpointConfig, &noRoute{})
	resources.Routes(custom, endpointConfig)
}

type noRoute struct{}

func (self *noRoute) Handle(ctx context.Context) (int, error) {
	return http.StatusNotFound, nil
}
