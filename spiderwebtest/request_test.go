package spiderwebtest

import (
	"net/http"
	"testing"

	"github.com/wspowell/spiderweb/examples/app"
	"github.com/wspowell/spiderweb/examples/app/mocks"
)

func Test_RouteTest(t *testing.T) {
	t.Parallel()

	server := app.SetupServer()

	TestRequest(t, server, GivenRequest(http.MethodPost, "/not_found").
		WithRequestBody("application/json", []byte(`{"my_string": "hello","my_int": 5}`)).
		ExpectResponse(http.StatusNotFound).
		WithEmptyBody())

	TestRequest(t, server, GivenRequest(http.MethodPost, "/resources").
		WithRequestBody("application/json", []byte(`{"my_string": "hello","my_int": 5}`)).
		ExpectResponse(http.StatusCreated).
		WithResponseBody("application/json", []byte(`{"output_string":"hello","output_int":5}`)))

	dbMock := &mocks.Datastore{}
	dbMock.On("RetrieveValue").Return("test")
	TestRequest(t, server, GivenRequest(http.MethodGet, "/resources/{id}").
		WithPathParam("id", "34").
		WithResourceMock("datastore", dbMock).
		ExpectResponse(http.StatusOK).
		WithResponseBody("application/json", []byte(`{"output_string":"test","output_int":34}`)))

	TestRequest(t, server, GivenRequest(http.MethodGet, "/resources/34").
		ExpectResponse(http.StatusInternalServerError).
		WithResponseBody("application/json", []byte(`{"code":"INTERNAL_ERROR","internal_code":"","message":"internal server error"}`)))
}
