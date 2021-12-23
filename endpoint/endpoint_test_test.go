package endpoint_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wspowell/context"
	"github.com/wspowell/errors"
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

// func (self *testPathParams) PathParameters() []request.Parameter {
// 	return []request.Parameter{
// 		request.Param[string]{
// 			Param: "param1",
// 			Value: &self.Param1,
// 		},
// 		request.Param[int]{
// 			Param: "param2",
// 			Value: &self.Param2,
// 		},
// 		request.Param[bool]{
// 			Param: "param3",
// 			Value: &self.Param3,
// 		},
// 	}
// }

type testQueryParams struct {
	Param1 string
	Param2 int
	Param3 bool
}

// func (self *testQueryParams) QueryParameters() []request.Parameter {
// 	return []request.Parameter{
// 		request.Param[string]{
// 			Param: "param1",
// 			Value: &self.Param1,
// 		},
// 		request.Param[int]{
// 			Param: "param2",
// 			Value: &self.Param2,
// 		},
// 		request.Param[bool]{
// 			Param: "param3",
// 			Value: &self.Param3,
// 		},
// 	}
// }

// func (self *Endpoint) PathParameters() []request.Parameter {
// 	return []request.Parameter{
// 		request.Param[string]{
// 			Param: "param1",
// 			Value: &self.Param1,
// 		},
// 		request.Param[int]{
// 			Param: "param2",
// 			Value: &self.Param2,
// 		},
// 		request.Param[bool]{
// 			Param: "param3",
// 			Value: &self.Param3,
// 		},
// 	}
// }

// type PathParam[T any] struct {
// 	Name  string
// 	Value T
// }

// func NewPathParam[T any](name string) PathParam[T] {
// 	var zero T
// 	return PathParam[T]{
// 		Name:  name,
// 		Value: zero,
// 	}
// }

// func (self *PathParam[T]) PathParameter() request.Parameter {
// 	return request.Param[T]{
// 		Param: self.Name,
// 		Value: &self.Value,
// 	}
// }

// type PathParamUserId      PathParam[string]("user_id")
// type PathParamResourceId  PathParam[int]("resource_id")

// type ResourceId request.Param[string]

// func (self *ResourceId) Name() string {
// 	return "resource_id"
// }

type Auth[T any] struct {
}

type foo struct {
	body.Request[requestBodyModel]
	body.Response[responseBodyModel]
	Param1 string
	Param2 int
	Param3 bool
}

func (self *foo) PathParameters() []request.Parameter {
	return []request.Parameter{
		request.NewParam("param1", &self.Param1),
		request.NewParam("param2", &self.Param2),
		request.NewParam("param3", &self.Param3),
	}
}

func (self *foo) Handle(ctx context.Context) (int, error) {
	// if self.Param1 != "value1" {
	// 	return httpstatus.InternalServerError, errors.New("param1 != value1")
	// }
	if self.Param2 != 13 {
		return httpstatus.InternalServerError, errors.New("param2 != 13")
	}
	if !self.Param3 {
		return httpstatus.InternalServerError, errors.New("param3 != true")
	}

	self.ResponseBody.OutputInt = 11
	self.ResponseBody.OutputString = "goodbye!"
	return httpstatus.OK, nil
}

func Test_foo(t *testing.T) {
	ctx := context.Background()

	e := foo{}
	eCopy := e

	request := testRequest(ctx)

	handle := handler.NewHandle(e)

	statusCode, responseBytes := handle.Run(ctx, request)
	fmt.Println(statusCode, string(responseBytes))

	statusCode, responseBytes = handle.Run(ctx, request)
	fmt.Println(statusCode, string(responseBytes))

	assert.Equal(t, eCopy, e)
	assert.Equal(t, httpstatus.OK, statusCode)
	assert.Equal(t, string(`{"outputString":"goodbye!","outputInt":11}`), string(responseBytes))
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

// func Test_test(t *testing.T) {
// 	ctx := context.Background()

// 	e := testEndpoint{
// 		Resources: Resources{},
// 	}

// 	eCopy := e

// 	registeredMimeTypes := map[string]mime.Handler{
// 		"application/json": &mime.Json{},
// 	}

// 	request := testRequest(ctx)

// 	executeHandler := handler.New(e)

// 	statusCode, responseBytes := executeHandler(ctx, request, registeredMimeTypes)
// 	fmt.Println(statusCode, string(responseBytes))

// 	statusCode, responseBytes = executeHandler(ctx, request, registeredMimeTypes)
// 	fmt.Println(statusCode, string(responseBytes))

// 	assert.Equal(t, eCopy, e)
// 	assert.Equal(t, httpstatus.OK, statusCode)
// 	assert.Equal(t, string(`{"outputString":"goodbye!","outputInt":11}`), string(responseBytes))
// }

// func Benchmark_test_test(b *testing.B) {
// 	ctx := context.Background()

// 	e := testEndpoint{
// 		Resources: Resources{},
// 	}

// 	registeredMimeTypes := map[string]mime.Handler{
// 		"application/json": &mime.Json{},
// 	}

// 	request := testRequest(ctx)

// 	run := handler.New(e)

// 	b.ResetTimer()

// 	for i := 0; i < b.N; i++ {
// 		statusCode, responseBytes := run(ctx, request, registeredMimeTypes)
// 		if statusCode == httpstatus.InternalServerError {
// 			panic(string(responseBytes))
// 		}
// 	}
// }
