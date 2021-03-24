package endpoint

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/wspowell/spiderweb/errors"
	"github.com/wspowell/spiderweb/logging"

	"github.com/valyala/fasthttp"
)

type errorResponse struct {
	InternalCode string `json:"internal_code"`
	Message      string `json:"message"`
}

type myErrorHandler struct{}

func (self myErrorHandler) MimeType() string {
	return mimeTypeJson
}

func (self myErrorHandler) HandleError(ctx *Context, httpStatus int, err error) (int, []byte) {
	if HasFrameworkError(err) {
		ctx.Error("internal error: %v", err)
		err = errors.New("AP0000", "internal server error")
	}

	responseBodyBytes, _ := json.Marshal(errorResponse{
		InternalCode: errors.InternalCode(err),
		Message:      err.Error(),
	})

	return httpStatus, responseBodyBytes
}

type myAuther struct{}

func (self myAuther) Auth(ctx *Context, headers map[string][]byte) (int, error) {
	for key, value := range headers {
		ctx.Info("%v:%v", key, string(value))
	}
	return http.StatusOK, nil
}

type myRequestValidator struct{}

func (self myRequestValidator) ValidateRequest(ctx *Context, requestBodyBytes []byte) (int, error) {
	return http.StatusOK, nil
}

type myResponseValidator struct{}

func (self myResponseValidator) ValidateResponse(ctx *Context, httpStatus int, responseBodyBytes []byte) (int, error) {
	return http.StatusOK, nil
}

// Fake database client to test setting resources.
type myDbClient struct {
	conn string
}

func (self *myDbClient) Conn() string {
	return self.conn
}

type myRequestBodyModel struct {
	MyString   string `json:"my_string"`
	MyInt      int    `json:"my_int"`
	ShouldFail bool   `json:"fail"`
}

type myResponseBodyModel struct {
	MyString string `json:"output_string"`
	MyInt    int    `json:"output_int"`
}

type myEndpoint struct {
	Test          string
	MyStringQuery string               `spiderweb:"query=id"`
	MyIntQuery    int                  `spiderweb:"query=num"`
	MyBoolQuery   bool                 `spiderweb:"query=flag"`
	MyStringParam string               `spiderweb:"path=id"`
	MyIntParam    int                  `spiderweb:"path=num"`
	MyFlagParam   bool                 `spiderweb:"path=flag"`
	MyDatabase    *myDbClient          `spiderweb:"resource=db"`
	RequestBody   *myRequestBodyModel  `spiderweb:"request,mime=application/json,validate"`
	ResponseBody  *myResponseBodyModel `spiderweb:"response,mime=application/json,validate"`
}

func (self *myEndpoint) Handle(ctx *Context) (int, error) {
	ctx.Debug("handling myEndpoint")

	if self.RequestBody.ShouldFail {
		return http.StatusUnprocessableEntity, errors.New("APP1234", "invalid input")
	}

	if self.MyStringQuery != "myid" {
		return http.StatusInternalServerError, errors.New("APP1111", "string query param not set")
	}

	if self.MyIntQuery != 13 {
		return http.StatusInternalServerError, errors.New("APP1111", "int query param not set")
	}

	if self.MyBoolQuery != true {
		return http.StatusInternalServerError, errors.New("APP1111", "bool query param not set")
	}

	if self.MyStringParam != "myid" {
		return http.StatusInternalServerError, errors.New("APP1111", "string path param not set")
	}

	if self.MyIntParam != 5 {
		return http.StatusInternalServerError, errors.New("APP1111", "int path param not set")
	}

	if self.MyFlagParam != true {
		return http.StatusInternalServerError, errors.New("APP1111", "bool path param not set")
	}

	if self.MyDatabase == nil {
		return http.StatusInternalServerError, errors.New("APP1111", "database not set")
	}

	if self.MyDatabase.Conn() != "myconnection" {
		return http.StatusInternalServerError, errors.New("APP1111", "database connection error")
	}

	self.ResponseBody = &myResponseBodyModel{
		MyString: self.RequestBody.MyString,
		MyInt:    self.RequestBody.MyInt,
	}

	return http.StatusOK, nil
}

func createTestEndpoint() *Endpoint {
	dbClient := myDbClient{
		conn: "myconnection",
	}

	config := Config{
		LogConfig:         logging.NewConfig(logging.LevelError, map[string]interface{}{}),
		ErrorHandler:      myErrorHandler{},
		Auther:            myAuther{},
		RequestValidator:  myRequestValidator{},
		ResponseValidator: myResponseValidator{},
		MimeTypeHandlers: map[string]MimeTypeHandler{
			"application/json": jsonHandler(),
		},
		Resources: map[string]ResourceFunc{
			"db": func() interface{} {
				return &dbClient
			},
		},
	}

	return NewEndpoint(config.Clone(), &myEndpoint{})
}

func newTestContext() *Context {
	var req fasthttp.Request

	//req.Header.SetMethod(method)
	req.Header.SetRequestURI("/resources/myid?id=myid&num=13&flag=true")
	req.Header.Set(fasthttp.HeaderHost, "localhost")
	req.Header.Set(fasthttp.HeaderUserAgent, "")
	req.Header.Set("Authorization", "auth-token")
	req.Header.SetContentType("application/json")
	req.Header.Set("Accept", "application/json")
	req.SetBody([]byte(`{"my_string": "hello", "my_int": 5}`))

	requestCtx := fasthttp.RequestCtx{}
	requestCtx.Init(&req, nil, nil)

	requestCtx.SetUserValue("id", "myid")
	requestCtx.SetUserValue("num", "5")
	requestCtx.SetUserValue("flag", "true")

	logConfig := logging.NewConfig(logging.LevelError, map[string]interface{}{})
	return NewContext(context.Background(), &requestCtx, logging.NewLogger(logConfig), 30*time.Second)
}

func Test_Endpoint_Default_Success(t *testing.T) {
	t.Parallel()

	endpoint := createTestEndpoint()
	ctx := newTestContext()

	httpStatus, responseBodyBytes := endpoint.Execute(ctx)

	if http.StatusOK != httpStatus {
		t.Errorf("expected HTTP status code to be %v, but got %v", http.StatusOK, httpStatus)
	}

	fmt.Println(string(responseBodyBytes))

	var responseBody myResponseBodyModel
	if err := json.Unmarshal(responseBodyBytes, &responseBody); err != nil {
		t.Errorf("failed to unmarshal test response: %v", err)
	}

	if "hello" != responseBody.MyString {
		t.Errorf("expected 'output_string' to be %v, but got %v", "hello", responseBody.MyString)
	}

	if 5 != responseBody.MyInt {
		t.Errorf("expected 'output_int' to be %v, but got %v", 5, responseBody.MyInt)
	}
}

func Test_Endpoint_Default_Error(t *testing.T) {
	t.Parallel()

	endpoint := createTestEndpoint()
	ctx := newTestContext()
	ctx.Request().SetBody([]byte(`{"my_string": "hello", "my_int": 5, "fail": true}`))
	httpStatus, responseBodyBytes := endpoint.Execute(ctx)

	if http.StatusUnprocessableEntity != httpStatus {
		t.Errorf("expected HTTP status code to be %v, but got %v", http.StatusOK, httpStatus)
	}

	fmt.Println(string(responseBodyBytes))

	var responseBody errorResponse
	if err := json.Unmarshal(responseBodyBytes, &responseBody); err != nil {
		t.Errorf("failed to unmarshal test response: %v", err)
	}

	if "APP1234" != responseBody.InternalCode {
		t.Errorf("expected 'internal_code' to be %v, but got %v", "APP1234", responseBody.InternalCode)
	}

	if "invalid input" != responseBody.Message {
		t.Errorf("expected 'message' to be %v, but got %v", "invalid input", responseBody.Message)
	}
}
