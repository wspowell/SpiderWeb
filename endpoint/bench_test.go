package endpoint_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"spiderweb/endpoint"
	"spiderweb/errors"
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
	RequestBody  *myRequestBodyModel  `spiderweb:"request,json,validate"`
	ResponseBody *myResponseBodyModel `spiderweb:"response,json,validate"`
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

func Benchmark_Endpoint_Default_Success(b *testing.B) {
	endpointRunner := endpoint.NewEndpoint(&myEndpoint{})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logConfig := logging.NewConfig(logging.LevelFatal, map[string]interface{}{})
			req := httptest.NewRequest(http.MethodGet, "/", bytes.NewBuffer([]byte(`{"my_string": "hello", "my_int": 5}`)))
			ctx := endpoint.NewContext(req, logging.NewLogger(logConfig))

			endpointRunner.Execute(ctx)
		}
	})
}

func Benchmark_Endpoint_Default_Error(b *testing.B) {
	endpointRunner := endpoint.NewEndpoint(&myEndpoint{})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logConfig := logging.NewConfig(logging.LevelFatal, map[string]interface{}{})
			req := httptest.NewRequest(http.MethodGet, "/", bytes.NewBuffer([]byte(`{"my_string": "hello", "my_int": 5, "fail": true}`)))
			ctx := endpoint.NewContext(req, logging.NewLogger(logConfig))

			endpointRunner.Execute(ctx)
		}
	})
}
