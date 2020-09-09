package spiderweb_test

import (
	"net/http"
	"testing"

	"spiderweb"
	"spiderweb/endpoint"
	"spiderweb/errors"
	"spiderweb/examples/auth"
	"spiderweb/examples/error_handlers"
	"spiderweb/examples/validators"
	"spiderweb/logging"
)

type myRequestBodyModel struct {
	MyString   string `json:"my_string"`
	MyInt      int    `json:"my_int"`
	ShouldFail bool   `json:"fail"`
}

type myResponseBodyModel struct {
	MyString string `json:"output_string"`
	MyInt    int    `json:"output_int"`
}

type myEndpoint struct {
	Test         string
	RequestBody  *myRequestBodyModel  `spiderweb:"request,mime=json,validate"`
	ResponseBody *myResponseBodyModel `spiderweb:"response,mime=json,validate"`
}

func (self *myEndpoint) Handle(ctx *endpoint.Context) (int, error) {
	ctx.Debug("handling myEndpoint")

	if self.RequestBody.ShouldFail {
		return http.StatusUnprocessableEntity, errors.New("APP1234", "invalid input")
	}

	self.ResponseBody = &myResponseBodyModel{
		MyString: self.RequestBody.MyString,
		MyInt:    self.RequestBody.MyInt,
	}

	return http.StatusOK, nil
}

func Test_Default_Endpoint_Config(t *testing.T) {
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

	router.Handle(http.MethodGet, "/", &myEndpoint{})
}
