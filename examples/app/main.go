package main

import (
	"net/http"

	"spiderweb"
	"spiderweb/endpoint"
	"spiderweb/examples/auth"
	"spiderweb/examples/error_handlers"
	"spiderweb/examples/validators"
	"spiderweb/logging"
)

func main() {
	endpointConfig := spiderweb.Config{
		EndpointConfig: endpoint.Config{
			Auther:            auth.Noop{},
			ErrorHandler:      error_handlers.ErrorJsonWithCodeResponse{},
			LogConfig:         logging.NewConfig(logging.LevelDebug, map[string]interface{}{}),
			MimeTypeHandlers:  map[string]endpoint.MimeTypeHandler{},
			RequestValidator:  validators.NoopRequest{},
			ResponseValidator: validators.NoopResponse{},
		},
		ServerHost: "localhost",
		ServerPort: 8080,
	}

	router := spiderweb.New(endpointConfig)

	router.Handle(http.MethodPost, "/resource", &postResource{})

	router.Run()
}
