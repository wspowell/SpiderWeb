package endpoint

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wspowell/context"
	"github.com/wspowell/errors"
	"github.com/wspowell/log"
)

type errorResponse struct {
	Message string `json:"message"`
}

type myErrorHandler struct{}

func (self myErrorHandler) HandleError(ctx *Context, httpStatus int, err error) (int, interface{}) {
	return httpStatus, errorResponse{
		Message: fmt.Sprintf("%#v", err),
	}
}

type myAuther struct{}

func (self myAuther) Auth(ctx *Context, VisitAllHeaders func(func(key, value []byte))) (int, error) {
	var statusCode int
	VisitAllHeaders(func(key, value []byte) {
		log.Info(ctx, "%v:%v", string(key), string(value))
	})

	return statusCode, nil
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

type Datastore interface {
	Conn() string
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
	MyDatabase    Datastore            `spiderweb:"resource=db"`
	RequestBody   *myRequestBodyModel  `spiderweb:"request,mime=application/json,validate"`
	ResponseBody  *myResponseBodyModel `spiderweb:"response,mime=application/json,validate"`
}

func (self *myEndpoint) Handle(ctx *Context) (int, error) {
	log.Debug(ctx, "handling myEndpoint")

	if self.RequestBody.ShouldFail {
		return http.StatusUnprocessableEntity, errors.New("APP1", "invalid input")
	}

	if self.MyStringQuery != "me" {
		return http.StatusInternalServerError, errors.New("APP2", "string query param not set")
	}

	if self.MyIntQuery != 13 {
		return http.StatusInternalServerError, errors.New("APP3", "int query param not set")
	}

	if self.MyBoolQuery != true {
		return http.StatusInternalServerError, errors.New("APP4", "bool query param not set")
	}

	if self.MyStringParam != "myid" {
		return http.StatusInternalServerError, errors.New("APP5", "string path param not set")
	}

	if self.MyIntParam != 5 {
		return http.StatusInternalServerError, errors.New("APP6", "int path param not set")
	}

	if self.MyFlagParam != true {
		return http.StatusInternalServerError, errors.New("APP7", "bool path param not set")
	}

	if self.MyDatabase == nil {
		return http.StatusInternalServerError, errors.New("APP8", "database not set")
	}

	if self.MyDatabase.Conn() != "myconnection" {
		return http.StatusInternalServerError, errors.New("APP9", "database connection error")
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

	config := &Config{
		LogConfig:         log.NewConfig(log.LevelError),
		ErrorHandler:      myErrorHandler{},
		Auther:            myAuther{},
		RequestValidator:  myRequestValidator{},
		ResponseValidator: myResponseValidator{},
		MimeTypeHandlers: map[string]*MimeTypeHandler{
			"application/json": jsonHandler(),
		},
		Resources: map[string]interface{}{
			"db": &dbClient,
		},
	}

	return NewEndpoint(config, &myEndpoint{})
}

func createDefaultTestEndpoint() *Endpoint {
	dbClient := myDbClient{
		conn: "myconnection",
	}

	config := &Config{
		Resources: map[string]interface{}{
			"db": &dbClient,
		},
	}

	return NewEndpoint(config, &myEndpoint{})
}

func Test_Endpoint_Success(t *testing.T) {
	t.Parallel()

	endpoint := createTestEndpoint()

	req, err := http.NewRequest(http.MethodPost, "/resources/myid/5/true?id=me&num=13&flag=true", strings.NewReader(`{"my_string": "hello", "my_int": 5}`))
	assert.Nil(t, err)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	requester := NewHttpRequester("/resources/{id}/{num}/{flag}", req)

	ctx := context.Local()
	ctx = log.WithContext(ctx, log.NewConfig(log.LevelError))
	var httpStatus int
	var responseBodyBytes []byte

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		endpointCtx := NewContext(ctx, requester)
		httpStatus, responseBodyBytes = endpoint.Execute(endpointCtx)
	}()
	wg.Wait()

	if http.StatusOK != httpStatus {
		t.Errorf("expected HTTP status code to be %v, but got %v", http.StatusOK, httpStatus)
	}

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

func Test_Endpoint_Default_Success(t *testing.T) {
	t.Parallel()

	endpoint := createDefaultTestEndpoint()

	req, err := http.NewRequest(http.MethodPost, "/resources/myid/5/true?id=me&num=13&flag=true", strings.NewReader(`{"my_string": "hello", "my_int": 5}`))
	assert.Nil(t, err)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	requester := NewHttpRequester("/resources/{id}/{num}/{flag}", req)

	ctx := context.Local()
	ctx = log.WithContext(ctx, log.NewConfig(log.LevelError))
	var httpStatus int
	var responseBodyBytes []byte

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		endpointCtx := NewContext(ctx, requester)
		httpStatus, responseBodyBytes = endpoint.Execute(endpointCtx)
	}()
	wg.Wait()

	if http.StatusOK != httpStatus {
		t.Errorf("expected HTTP status code to be %v, but got %v", http.StatusOK, httpStatus)
	}

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

func Test_Endpoint_Error(t *testing.T) {
	t.Parallel()

	endpoint := createTestEndpoint()

	req, err := http.NewRequest(http.MethodPost, "/resources/myid/5/true?id=me&num=13&flag=true", strings.NewReader(`{"my_string": "hello", "my_int": 5, "fail": true}`))
	assert.Nil(t, err)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	requester := NewHttpRequester("/resources/{id}/{num}/{flag}", req)

	ctx := context.Local()
	ctx = log.WithContext(ctx, log.NewConfig(log.LevelError))
	var httpStatus int
	var responseBodyBytes []byte

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		endpointCtx := NewContext(ctx, requester)
		httpStatus, responseBodyBytes = endpoint.Execute(endpointCtx)
	}()
	wg.Wait()

	if http.StatusUnprocessableEntity != httpStatus {
		t.Errorf("expected HTTP status code to be %v, but got %v", http.StatusOK, httpStatus)
	}

	var responseBody errorResponse
	if err := json.Unmarshal(responseBodyBytes, &responseBody); err != nil {
		t.Errorf("failed to unmarshal test response: %v", err)
	}

	if "[APP1] invalid input" != responseBody.Message {
		t.Errorf("expected 'message' to be '%v', but got '%v'", "[APP1] invalid input", responseBody.Message)
	}
}

func Test_Endpoint_Default_Error(t *testing.T) {
	t.Parallel()

	endpoint := createDefaultTestEndpoint()

	req, err := http.NewRequest(http.MethodPost, "/resources/myid/5/true?id=me&num=13&flag=true", strings.NewReader(`{"my_string": "hello", "my_int": 5, "fail": true}`))
	assert.Nil(t, err)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	requester := NewHttpRequester("/resources/{id}/{num}/{flag}", req)

	ctx := context.Local()
	ctx = log.WithContext(ctx, log.NewConfig(log.LevelError))
	var httpStatus int
	var responseBodyBytes []byte

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		endpointCtx := NewContext(ctx, requester)
		httpStatus, responseBodyBytes = endpoint.Execute(endpointCtx)
	}()
	wg.Wait()

	if http.StatusUnprocessableEntity != httpStatus {
		t.Errorf("expected HTTP status code to be %v, but got %v", http.StatusOK, httpStatus)
	}

	if `{"message":"[APP1] invalid input"}` != string(responseBodyBytes) {
		t.Errorf("expected 'message' to be '%v', but got '%v'", "[APP1] invalid input", string(responseBodyBytes))
	}
}
