package spiderweb_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/wspowell/errors"
	"github.com/wspowell/logging"
	"github.com/wspowell/spiderweb"
	"github.com/wspowell/spiderweb/endpoint"
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
	serverConfig := &spiderweb.ServerConfig{
		LogConfig:    logging.NewConfig(logging.LevelDebug, map[string]interface{}{}),
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	server := spiderweb.NewServer(serverConfig)

	server.Handle(&endpoint.Config{}, http.MethodGet, "/", &myEndpoint{})

	// TODO: Add some checks.
}
