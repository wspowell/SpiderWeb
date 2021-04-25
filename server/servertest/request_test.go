package servertest

import (
	"net/http"
	"testing"
)

func Test_RouteTest(t *testing.T) {
	t.Parallel()

	sample := routes()

	TestRequest(t, sample, "Route not found",
		GivenRequest(http.MethodPost, "/not_found").
			WithRequestBody("application/json", []byte(`{"my_string": "hello","my_int": 5}`)).
			ExpectResponse(http.StatusNotFound).
			WithEmptyBody())

	TestRequest(t, sample, "Success POST /sample",
		GivenRequest(http.MethodPost, "/sample").
			WithRequestBody("application/json", []byte(`{"my_string": "hello","my_int": 5}`)).
			ExpectResponse(http.StatusCreated).
			WithResponseBody("application/json", []byte(`{"output_string":"hello","output_int":5}`)))

	dbMock := &mockDatastore{}
	dbMock.On("RetrieveValue").Return("test")
	TestRequest(t, sample, "Success GET /sample/{id}",
		GivenRequest(http.MethodGet, "/sample/{id}").
			WithPathParam("id", "34").
			WithResourceMock("datastore", dbMock).
			ExpectResponse(http.StatusOK).
			WithResponseBody("application/json", []byte(`{"output_string":"test","output_int":34}`)))

	// Not mocked, so it returns 500.
	TestRequest(t, sample, "Failure, not mocked",
		GivenRequest(http.MethodGet, "/sample/{id}").
			ExpectResponse(http.StatusInternalServerError).
			WithResponseBody("application/json", []byte(`{"message":"[SW000] internal server error"}`)))
}
