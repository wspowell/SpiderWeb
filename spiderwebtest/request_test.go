package spiderwebtest

import (
	"net/http"
	"testing"

	"github.com/wspowell/spiderweb/examples/app"
)

func Test_RouteTest(t *testing.T) {
	t.Parallel()

	server := app.SetupServer()

	TestRequest(t, server, GivenRequest(http.MethodPost, "/not_found").
		WithRequestBody("application/json", []byte(`{"my_string": "hello","my_int": 5}`)).
		Expect(http.StatusNotFound, "application/json", []byte(``)))

	TestRequest(t, server, GivenRequest(http.MethodPost, "/resources").
		WithRequestBody("application/json", []byte(`{"my_string": "hello","my_int": 5}`)).
		Expect(http.StatusCreated, "application/json", []byte(`{"output_string":"hello","output_int":5}`)))

	TestRequest(t, server, GivenRequest(http.MethodGet, "/resources/34").
		Expect(http.StatusOK, "application/json", []byte(`{"output_string":"test","output_int":34}`)))
}
