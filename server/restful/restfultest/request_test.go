package restfultest_test

import (
	"fmt"
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
		GivenRequest().
		Body([]byte(`{"myString": "hello","myInt": 5}`)).
		ExpectResponse().
		Header(httpheader.ContentType, "text/plain; charset=utf-8").
		StatusCode(http.StatusNotFound).
		Run(t)
}

func Test_RouteNotFound_WithAccept(t *testing.T) {
	t.Parallel()

	dbMock := &test.MockDatastore{}
	defer dbMock.AssertExpectations(t)

	restfultest.Server(RoutesTest(dbMock)).
		Case("Route not found").
		ForRoute(http.MethodPost, "/not_found").
		GivenRequest().
		Header(httpheader.Accept, "application/json").
		Body([]byte(`{"myString": "hello","myInt": 5}`)).
		ExpectResponse().
		Header(httpheader.ContentType, "application/json").
		StatusCode(http.StatusNotFound).
		Run(t)
}

func Test_POST(t *testing.T) {
	t.Parallel()

	restfultest.Server(RoutesTest(&test.Database{})).
		Case("Success POST /sample").
		ForRoute(http.MethodPost, "/sample").
		GivenRequest().
		Body([]byte(`{"myString": "hello","myInt": 5}`)).
		Header(httpheader.ContentType, "application/json").
		Header(httpheader.Accept, "application/json").
		ExpectResponse().
		Header(httpheader.ContentType, "application/json").
		Body([]byte(`{"outputString":"hello","outputInt":5}`)).
		StatusCode(http.StatusCreated).
		Run(t)
}

func Test_POST_Mocked(t *testing.T) {
	t.Parallel()

	dbMock := &test.MockDatastore{}
	dbMock.On("RetrieveValue").Return("test")
	defer dbMock.AssertExpectations(t)

	fmt.Printf("%p\n", dbMock)

	restfultest.Server(RoutesTest(dbMock)).
		Case("Success GET /sample/{id}").
		ForRoute(http.MethodGet, "/sample/34").
		GivenRequest().
		Header(httpheader.ContentType, "application/json").
		Header(httpheader.Accept, "application/json").
		ExpectResponse().
		Header(httpheader.ContentType, "application/json").
		Body([]byte(`{"outputString":"test","outputInt":34}`)).
		StatusCode(http.StatusOK).
		Run(t)
}

func Test_POST_NotMocked(t *testing.T) {
	t.Parallel()

	dbMock := &test.MockDatastore{}
	defer dbMock.AssertExpectations(t)

	// Not mocked, so it returns 500.
	restfultest.Server(RoutesTest(dbMock)).
		Case("Failure, not mocked").
		ForRoute(http.MethodGet, "/sample/34").
		GivenRequest().
		Header(httpheader.ContentType, "application/json").
		Header(httpheader.Accept, "application/json").
		ExpectResponse().
		Header(httpheader.ContentType, "application/json").
		Body([]byte(`{"error":"internal server error"}`)).
		StatusCode(http.StatusInternalServerError).
		Run(t)
}
