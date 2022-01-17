package restfultest_test

import (
	"net/http"
	"testing"

	"github.com/wspowell/spiderweb/httpheader"
	"github.com/wspowell/spiderweb/server/restful/restfultest"
	"github.com/wspowell/spiderweb/test"
)

func Test_RouteNotFound(t *testing.T) {
	t.Parallel()

	dbMock := &test.MockDatastore{}
	defer dbMock.AssertExpectations(t)

	restfultest.Server(RoutesTest(dbMock)).
		Case("Route not found").
		ForRoute(http.MethodPost, "/not_found").
		WithRequestBody([]byte(`{"myString": "hello","myInt": 5}`)).
		WithHeader(httpheader.ContentType, "application/json").
		WithHeader(httpheader.Accept, "application/json").
		ExpectStatusCode(http.StatusNotFound).
		Run(t)
}

func Test_POST_sample(t *testing.T) {
	t.Parallel()

	dbMock := &test.MockDatastore{}
	defer dbMock.AssertExpectations(t)

	restfultest.Server(RoutesTest(dbMock)).
		Case("Success POST /sample").
		ForRoute(http.MethodPost, "/sample").
		WithRequestBody([]byte(`{"myString": "hello","myInt": 5}`)).
		WithHeader(httpheader.ContentType, "application/json").
		WithHeader(httpheader.Accept, "application/json").
		ExpectStatusCode(http.StatusCreated).
		WithResponseBody([]byte(`{"outputString":"hello","outputInt":5}`)).
		WithHeader(httpheader.ContentType, "application/json").
		Run(t)
}

func Test_POST_sample_id_34(t *testing.T) {
	t.Parallel()

	dbMock := &test.MockDatastore{}
	dbMock.On("RetrieveValue").Return("test")
	defer dbMock.AssertExpectations(t)

	restfultest.Server(RoutesTest(dbMock)).
		Case("Success GET /sample/{id}").
		ForRoute(http.MethodGet, "/sample/{id}").
		WithHeader(httpheader.ContentType, "application/json").
		WithHeader(httpheader.Accept, "application/json").
		WithPathParam("id", "34").
		ExpectStatusCode(http.StatusOK).
		WithResponseBody([]byte(`{"outputString":"test","outputInt":34}`)).
		WithHeader(httpheader.ContentType, "application/json").
		Run(t)
}

func Test_resource_not_mocked(t *testing.T) {
	t.Parallel()

	dbMock := &test.MockDatastore{}
	defer dbMock.AssertExpectations(t)

	// Not mocked, so it returns 500.
	restfultest.Server(RoutesTest(dbMock)).
		Case("Failure, not mocked").
		ForRoute(http.MethodGet, "/sample/{id}").
		WithPathParam("id", "34").
		WithHeader(httpheader.ContentType, "application/json").
		WithHeader(httpheader.Accept, "application/json").
		ExpectStatusCode(http.StatusInternalServerError).
		WithResponseBody([]byte(`{"error":"internal server error"}`)).
		WithHeader(httpheader.ContentType, "application/json").
		Run(t)
}
