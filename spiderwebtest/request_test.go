package spiderwebtest

import (
	"net/http"
	"testing"

	"spiderweb/examples/app"
)

func Test_EndpointTest(t *testing.T) {
	server := app.SetupServer()
	TestRequest(t, server, GivenRequest(http.MethodPost, "/resources").
		WithRequestBody([]byte(`{"my_string": "hello","my_int": 5}`)).
		Expect(http.StatusCreated, []byte(`{"output_string":"hello","output_int":5}`)))

	TestRequest(t, server, GivenRequest(http.MethodGet, "/resources/34").
		Expect(http.StatusOK, []byte(`{"output_string":"test","output_int":34}`)))
}
