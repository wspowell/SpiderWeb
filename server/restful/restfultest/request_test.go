package restfultest_test

import (
	"net/http"
	"testing"

	"github.com/wspowell/spiderweb/server/restful/restfultest"
	"github.com/wspowell/spiderweb/test"
)

func Test_RouteNotFound(t *testing.T) {
	t.Parallel()

	restfultest.TestCase(RoutesTest(nil), "Route not found").
		GivenRequest(http.MethodPost, "/not_found").
		WithRequestBody("application/json", []byte(`{"myString": "hello","myInt": 5}`)).
		ExpectResponse(http.StatusNotFound).
		WithEmptyBody().
		Run(t)
}

func Test_POST_sample(t *testing.T) {
	t.Parallel()

	restfultest.TestCase(RoutesTest(nil), "Success POST /sample").
		GivenRequest(http.MethodPost, "/sample").
		WithRequestBody("application/json", []byte(`{"myString": "hello","myInt": 5}`)).
		ExpectResponse(http.StatusCreated).
		WithResponseBody("application/json", []byte(`{"outputString":"hello","outputInt":5}`)).
		Run(t)
}

func Test_POST_sample_id_34(t *testing.T) {
	t.Parallel()

	dbMock := &test.MockDatastore{}
	dbMock.On("RetrieveValue").Return("test")
	restfultest.TestCase(RoutesTest(dbMock), "Success GET /sample/{id}").
		GivenRequest(http.MethodGet, "/sample/{id}").
		WithPathParam("id", "34").
		WithResourceMock("datastore", dbMock).
		ExpectResponse(http.StatusOK).
		WithResponseBody("application/json", []byte(`{"outputString":"test","outputInt":34}`)).
		Run(t)
}

func Test_resource_not_mocked(t *testing.T) {
	t.Parallel()

	// Not mocked, so it returns 500.
	restfultest.TestCase(RoutesTest(nil), "Failure, not mocked").
		GivenRequest(http.MethodGet, "/sample/{id}").
		WithPathParam("id", "34").
		ExpectResponse(http.StatusInternalServerError).
		WithResponseBody("application/json", []byte(`{"error":"internal server error"}`)).
		Run(t)
}
