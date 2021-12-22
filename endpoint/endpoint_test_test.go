package endpoint_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wspowell/context"
	"github.com/wspowell/spiderweb/body"
	"github.com/wspowell/spiderweb/endpoint"
	"github.com/wspowell/spiderweb/handler"
	"github.com/wspowell/spiderweb/httpheader"
	"github.com/wspowell/spiderweb/httpstatus"
	"github.com/wspowell/spiderweb/mime"
	"github.com/wspowell/spiderweb/request"
)

type testUser struct {
	userData string
}

type requestBodyModel struct {
	mime.Json

	MyString   string `json:"myString"`
	MyInt      int    `json:"myInt"`
	ShouldFail bool   `json:"shouldFail"`
}

type responseBodyModel struct {
	mime.Json

	OutputString string `json:"outputString"`
	OutputInt    int    `json:"outputInt"`
}

type MySql struct {
}

func (self *MySql) Get() *MySql {
	return self
}

type Resources struct {
	MySql1 Datastore
	MySql2 Datastore
	Redis  Datastore
}

type testPathParams struct {
	Param1 string
	Param2 int
	Param3 bool
}

func (self *testPathParams) PathParameters() []request.Parameter {
	return []request.Parameter{
		request.Param[string]{
			Param: "param1",
			Value: &self.Param1,
		},
		request.Param[int]{
			Param: "param2",
			Value: &self.Param2,
		},
		request.Param[bool]{
			Param: "param3",
			Value: &self.Param3,
		},
	}
}

type testQueryParams struct {
	Param1 string
	Param2 int
	Param3 bool
}

func (self *testQueryParams) QueryParameters() []request.Parameter {
	return []request.Parameter{
		request.Param[string]{
			Param: "param1",
			Value: &self.Param1,
		},
		request.Param[int]{
			Param: "param2",
			Value: &self.Param2,
		},
		request.Param[bool]{
			Param: "param3",
			Value: &self.Param3,
		},
	}
}

type testEndpoint struct {
	Resources

	body.Request[requestBodyModel]
	body.Response[responseBodyModel]

	testPathParams
	testQueryParams
}

func (self *testEndpoint) Handle(ctx context.Context) (int, error) {
	self.ResponseBody.OutputInt = 11
	self.ResponseBody.OutputString = "goodbye!"
	return httpstatus.OK, nil
}

func testRequest(ctx context.Context) endpoint.Requester {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/resources/value1/13/true?param1=me&param2=5&param3=true", strings.NewReader(`{"myString": "hello", "myInt": 5}`))
	if err != nil {
		panic(err)
	}

	req.Header.Add(httpheader.ContentType, "application/json")
	req.Header.Add(httpheader.Accept, "application/json")
	req.Header.Add(httpheader.Authorization, "valid-token")

	requester, err := endpoint.NewHttpRequester("/resources/{param1}/{param2}/{param3}", req)
	if err != nil {
		panic(err)
	}

	return requester
}

func Test_test(t *testing.T) {
	ctx := context.Background()

	e := testEndpoint{
		Resources: Resources{},
	}

	eCopy := e

	registeredMimeTypes := map[string]mime.Handler{
		"application/json": &mime.Json{},
	}

	request := testRequest(ctx)

	executeHandler := handler.New(e)

	statusCode, responseBytes := executeHandler(ctx, request, registeredMimeTypes)
	fmt.Println(statusCode, string(responseBytes))

	statusCode, responseBytes = executeHandler(ctx, request, registeredMimeTypes)
	fmt.Println(statusCode, string(responseBytes))

	assert.Equal(t, eCopy, e)
	assert.Equal(t, httpstatus.OK, statusCode)
	assert.Equal(t, string(`{"outputString":"goodbye!","outputInt":11}`), string(responseBytes))
}

func Benchmark_test_test(b *testing.B) {
	ctx := context.Background()

	e := testEndpoint{
		Resources: Resources{},
	}

	registeredMimeTypes := map[string]mime.Handler{
		"application/json": &mime.Json{},
	}

	request := testRequest(ctx)

	run := handler.New(e)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		statusCode, responseBytes := run(ctx, request, registeredMimeTypes)
		if statusCode == httpstatus.InternalServerError {
			panic(string(responseBytes))
		}
	}
}
