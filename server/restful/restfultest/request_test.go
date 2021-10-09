package restfultest_test

import (
	"net/http"
	"testing"

	"github.com/wspowell/spiderweb/server/restful/restfultest"
	"github.com/wspowell/spiderweb/test"
)

func Test_RouteTest(t *testing.T) {
	t.Parallel()

	sample := Routes()

	restfultest.TestCase(sample, "Route not found").
		GivenRequest(http.MethodPost, "/not_found").
		WithRequestBody("application/json", []byte(`{"my_string": "hello","my_int": 5}`)).
		ExpectResponse(http.StatusNotFound).
		WithEmptyBody().
		RunParallel(t)

	restfultest.TestCase(sample, "Success POST /sample").
		GivenRequest(http.MethodPost, "/sample").
		WithRequestBody("application/json", []byte(`{"my_string": "hello","my_int": 5}`)).
		ExpectResponse(http.StatusCreated).
		WithResponseBody("application/json", []byte(`{"output_string":"hello","output_int":5}`)).
		RunParallel(t)

	dbMock := &test.MockDatastore{}
	dbMock.On("RetrieveValue").Return("test")
	restfultest.TestCase(sample, "Success GET /sample/{id}").
		GivenRequest(http.MethodGet, "/sample/{id}").
		WithPathParam("id", "34").
		WithResourceMock("datastore", dbMock).
		ExpectResponse(http.StatusOK).
		WithResponseBody("application/json", []byte(`{"output_string":"test","output_int":34}`)).
		RunParallel(t)

	// Not mocked, so it returns 500.
	restfultest.TestCase(sample, "Failure, not mocked").
		GivenRequest(http.MethodGet, "/sample/{id}").
		ExpectResponse(http.StatusInternalServerError).
		WithResponseBody("application/json", []byte(`{"message":"[SW001] internal server error"}`)).
		RunParallel(t)
}
