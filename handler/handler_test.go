package handler_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wspowell/context"
	"github.com/wspowell/errors"
	"github.com/wspowell/spiderweb/body"
	"github.com/wspowell/spiderweb/handler"
	"github.com/wspowell/spiderweb/httpheader"
	"github.com/wspowell/spiderweb/httpstatus"
	"github.com/wspowell/spiderweb/httptrip"
	"github.com/wspowell/spiderweb/mime"
	"github.com/wspowell/spiderweb/request"
)

func testRequest(ctx context.Context) *httptrip.HttpRoundTrip {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/resources/myid/5/true?id=me&num=5&flag=true", strings.NewReader(`{"myString": "hello", "myInt": 5}`))
	if err != nil {
		panic(err)
	}

	req.Header.Add(httpheader.ContentType, "application/json")
	req.Header.Add(httpheader.Accept, "application/json")
	req.Header.Add(httpheader.Authorization, "valid-token")

	reqRes, err := httptrip.NewHttpRoundTrip("/resources/{id}/{num}/{flag}", req)
	if err != nil {
		panic(err)
	}

	return reqRes
}

type Auth[T any] struct {
}

type RequestBodyModel struct {
	mime.Json

	MyString   string `json:"myString"`
	MyInt      int    `json:"myInt"`
	ShouldFail bool   `json:"shouldFail"`
}

type ResponseBodyModel struct {
	mime.Json

	OutputString string `json:"outputString"`
	OutputInt    int    `json:"outputInt"`
}

type testHandler struct {
	body.Request[RequestBodyModel]
	body.Response[ResponseBodyModel]
	Param1 string
	Param2 int
	Param3 bool
}

func (self *testHandler) Handle(ctx context.Context) (int, error) {
	if self.Param1 != "myid" {
		return httpstatus.InternalServerError, errors.New("param1 != value1")
	}
	if self.Param2 != 5 {
		return httpstatus.InternalServerError, errors.New("param2 != 13")
	}
	if !self.Param3 {
		return httpstatus.BadRequest, errors.New("param3 != true")
	}

	self.ResponseBody.OutputInt = 11
	self.ResponseBody.OutputString = "goodbye!"
	return httpstatus.OK, nil
}

func (self *testHandler) PathParameters() []request.Parameter {
	return []request.Parameter{
		request.NewParam("id", &self.Param1),
		request.NewParam("num", &self.Param2),
		request.NewParam("flag", &self.Param3),
	}
}

func Test_foo(t *testing.T) {
	ctx := context.Background()

	handlerStruct := testHandler{}
	handlerStructCopy := handlerStruct

	reqRes := testRequest(ctx)

	handle := handler.NewHandle(handlerStruct)

	handle.Runner().Run(ctx, reqRes)
	fmt.Println(reqRes.StatusCode(), string(reqRes.ResponseBody()))

	reqRes.Close()
	reqRes = testRequest(ctx)

	handle.Runner().Run(ctx, reqRes)
	fmt.Println(reqRes.StatusCode(), string(reqRes.ResponseBody()))

	assert.Equal(t, handlerStructCopy, handlerStruct)
	assert.Equal(t, httpstatus.OK, reqRes.StatusCode())
	assert.Equal(t, string(`{"outputString":"goodbye!","outputInt":11}`), string(reqRes.ResponseBody()))

	reqRes.Close()
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