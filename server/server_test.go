package server_test

import (
	"net/http"
	"testing"

	"github.com/wspowell/context"
	"github.com/wspowell/errors"
	"github.com/wspowell/log"
	"github.com/wspowell/spiderweb/endpoint"
	"github.com/wspowell/spiderweb/server"
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

func (self *myEndpoint) Handle(ctx context.Context) (int, error) {
	log.Debug(ctx, "handling myEndpoint")

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
	server := server.New(nil)

	server.Handle(&endpoint.Config{}, http.MethodGet, "/", &myEndpoint{})

	// TODO: Add some checks.
}
