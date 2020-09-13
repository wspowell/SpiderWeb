package app

import (
	"net/http"
	"time"

	"spiderweb"
	"spiderweb/endpoint"
	"spiderweb/examples/auth"
	"spiderweb/examples/error_handlers"
	"spiderweb/examples/validators"
	"spiderweb/logging"
)

func main() {
	server := SetupServer()
	server.Listen()
}

func SetupServer() spiderweb.Server {

	serverConfig := spiderweb.NewServerConfig("localhost", 8080, endpoint.Config{
		Auther:            auth.Noop{},
		ErrorHandler:      error_handlers.ErrorJsonWithCodeResponse{},
		LogConfig:         logging.NewConfig(logging.LevelDebug, map[string]interface{}{}),
		MimeTypeHandlers:  map[string]endpoint.MimeTypeHandler{},
		RequestValidator:  validators.NoopRequest{},
		ResponseValidator: validators.NoopResponse{},
		Timeout:           30 * time.Second,
	})

	serverConfig.Handle(http.MethodPost, "/resources", &PostResource{})
	serverConfig.Handle(http.MethodGet, "/resources/{id}", &GetResource{})

	return spiderweb.NewServer(serverConfig)
}
