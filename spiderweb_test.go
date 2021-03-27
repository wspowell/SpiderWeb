package spiderweb_test

import (
	"net/http"
	"testing"

	"github.com/wspowell/errors"
	"github.com/wspowell/spiderweb"
	"github.com/wspowell/spiderweb/endpoint"
	"github.com/wspowell/spiderweb/examples/auth"
	"github.com/wspowell/spiderweb/examples/error_handlers"
	"github.com/wspowell/spiderweb/examples/validators"
	"github.com/wspowell/spiderweb/logging"
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
	RequestBody  *myRequestBodyModel  `spiderweb:"request,mime=application/json,validate"`
	ResponseBody *myResponseBodyModel `spiderweb:"response,mime=application/json,validate"`
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

func Test_Default_Server_Config(t *testing.T) {
	serverConfig := spiderweb.NewServerConfig("localhost", 8080, endpoint.Config{
		Auther:            auth.Noop{},
		ErrorHandler:      error_handlers.ErrorJsonWithCodeResponse{},
		LogConfig:         logging.NewConfig(logging.LevelDebug, map[string]interface{}{}),
		MimeTypeHandlers:  map[string]endpoint.MimeTypeHandler{},
		RequestValidator:  validators.NoopRequest{},
		ResponseValidator: validators.NoopResponse{},
	})

	serverConfig.Handle(http.MethodGet, "/", &myEndpoint{})

	spiderweb.NewServer(serverConfig)

	// TODO: Add some checks.
}
