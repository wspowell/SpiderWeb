package endpoint_test

import (
	"bytes"
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

	"github.com/wspowell/spiderweb/endpoint"
	"github.com/wspowell/spiderweb/httpheader"
	"github.com/wspowell/spiderweb/httpstatus"
)

type errorResponse struct {
	Message string `json:"message"`
}

type myErrorHandler struct{}

func (self myErrorHandler) HandleError(ctx context.Context, httpStatus int, err error) (int, interface{}) {
	return httpStatus, errorResponse{
		Message: fmt.Sprintf("%v", err),
	}
}

type user struct {
	userData string
}

func (self *user) Authorization(ctx context.Context, peekHeader func(key string) []byte) (int, error) {
	var statusCode int

	if !bytes.EqualFold([]byte("valid-token"), peekHeader(httpheader.Authorization)) {
		return httpstatus.Unauthorized, errors.New("auth", "invalid auth token")
	}

	self.userData = "myUserData"

	return statusCode, nil
}

type myRequestValidator struct{}

func (self myRequestValidator) ValidateRequest(ctx context.Context, requestBodyBytes []byte) (int, error) {
	return httpstatus.OK, nil
}

type myResponseValidator struct{}

func (self myResponseValidator) ValidateResponse(ctx context.Context, httpStatus int, responseBodyBytes []byte) (int, error) {
	return httpstatus.OK, nil
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
	MyString   string `json:"myString"`
	MyInt      int    `json:"myInt"`
	ShouldFail bool   `json:"shouldFail"`
}

type myResponseBodyModel struct {
	OutputString string `json:"outputString"`
	OutputInt    int    `json:"outputInt"`
}

type myEndpoint struct {
	Test            string
	User            *user                `spiderweb:"auth"`
	MyStringQuery   string               `spiderweb:"query=id,required"`
	MyIntQuery      int                  `spiderweb:"query=num"`
	MyBoolQuery     bool                 `spiderweb:"query=flag"`
	MyOptionalQuery string               `spiderweb:"query=optional"`
	MyStringParam   string               `spiderweb:"path=id"`
	MyIntParam      int                  `spiderweb:"path=num"`
	MyFlagParam     bool                 `spiderweb:"path=flag"`
	MyDatabase      Datastore            `spiderweb:"resource=db"`
	RequestBody     *myRequestBodyModel  `spiderweb:"request,mime=application/json,validate"`
	ResponseBody    *myResponseBodyModel `spiderweb:"response,mime=application/json,validate"`
}

func (self *myEndpoint) Handle(ctx context.Context) (int, error) {
	log.Debug(ctx, "handling myEndpoint")

	if self.User.userData != "myUserData" {
		return httpstatus.Forbidden, errors.New("APP0", "invalid auth")
	}

	if self.RequestBody.ShouldFail {
		return httpstatus.UnprocessableEntity, errors.New("APP1", "invalid input")
	}

	if self.MyStringQuery != "me" {
		return httpstatus.InternalServerError, errors.New("APP2", "string query param not set")
	}

	if self.MyIntQuery != 13 {
		return httpstatus.InternalServerError, errors.New("APP3", "int query param not set")
	}

	if self.MyBoolQuery != true {
		return httpstatus.InternalServerError, errors.New("APP4", "bool query param not set")
	}

	if self.MyStringParam != "myid" {
		return httpstatus.InternalServerError, errors.New("APP5", "string path param not set")
	}

	if self.MyIntParam != 5 {
		return httpstatus.InternalServerError, errors.New("APP6", "int path param not set")
	}

	if self.MyFlagParam != true {
		return httpstatus.InternalServerError, errors.New("APP7", "bool path param not set")
	}

	if self.MyDatabase == nil {
		return httpstatus.InternalServerError, errors.New("APP8", "database not set")
	}

	if self.MyDatabase.Conn() != "myconnection" {
		return httpstatus.InternalServerError, errors.New("APP9", "database connection error")
	}

	if self.MyOptionalQuery != "" {
		return httpstatus.InternalServerError, errors.New("APP10", "optional path param shoulde not be set")
	}

	self.ResponseBody = &myResponseBodyModel{
		OutputString: self.RequestBody.MyString,
		OutputInt:    self.RequestBody.MyInt,
	}

	return httpstatus.OK, nil
}

func createTestEndpoint() *endpoint.Endpoint {
	ctx := context.Local()
	log.WithContext(ctx, log.NewConfig(log.LevelError))

	dbClient := myDbClient{
		conn: "myconnection",
	}

	config := &endpoint.Config{
		LogConfig:         log.NewConfig(log.LevelError),
		ErrorHandler:      myErrorHandler{},
		RequestValidator:  myRequestValidator{},
		ResponseValidator: myResponseValidator{},
		MimeTypeHandlers: map[string]*endpoint.MimeTypeHandler{
			"application/json": endpoint.JsonHandler(),
		},
		Resources: map[string]interface{}{
			"db": &dbClient,
		},
	}

	return endpoint.NewEndpoint(ctx, config, &myEndpoint{})
}

func createDefaultTestEndpoint() *endpoint.Endpoint {
	ctx := context.Local()
	log.WithContext(ctx, log.NewConfig(log.LevelError))

	dbClient := myDbClient{
		conn: "myconnection",
	}

	config := &endpoint.Config{
		Resources: map[string]interface{}{
			"db": &dbClient,
		},
	}

	return endpoint.NewEndpoint(ctx, config, &myEndpoint{})
}

func Test_Endpoint_Success(t *testing.T) {
	t.Parallel()

	testEndpoint := createTestEndpoint()
	checkSuccessCase(t, testEndpoint)
}

func Test_Endpoint_Default_Success(t *testing.T) {
	t.Parallel()

	testEndpoint := createDefaultTestEndpoint()
	checkSuccessCase(t, testEndpoint)
}

func checkSuccessCase(t *testing.T, testEndpoint *endpoint.Endpoint) {
	t.Helper()

	ctx := context.Local()
	ctx = log.WithContext(ctx, log.NewConfig(log.LevelError))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/resources/myid/5/true?id=me&num=13&flag=true", strings.NewReader(`{"myString": "hello", "myInt": 5}`))
	assert.Nil(t, err)

	req.Header.Add(httpheader.ContentType, "application/json")
	req.Header.Add(httpheader.Accept, "application/json")
	req.Header.Add(httpheader.Authorization, "valid-token")

	requester, err := endpoint.NewHttpRequester("/resources/{id}/{num}/{flag}", req)
	assert.Nil(t, err)

	var httpStatus int
	var responseBodyBytes []byte

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		httpStatus, responseBodyBytes = testEndpoint.Execute(ctx, requester)
	}()
	wg.Wait()

	if httpstatus.OK != httpStatus {
		t.Errorf("expected HTTP status code to be %v, but got %v", httpstatus.OK, httpStatus)
	}

	var responseBody myResponseBodyModel
	if err := json.Unmarshal(responseBodyBytes, &responseBody); err != nil {
		t.Errorf("failed to unmarshal test response: %v", err)
	}

	if responseBody.OutputString != "hello" {
		t.Errorf("expected 'outputString' to be %v, but got %v", "hello", responseBody.OutputString)
	}

	if responseBody.OutputInt != 5 {
		t.Errorf("expected 'outputInt' to be %v, but got %v", 5, responseBody.OutputInt)
	}
}

func Test_Endpoint_Error(t *testing.T) {
	t.Parallel()

	testEndpoint := createTestEndpoint()

	ctx := context.Local()
	ctx = log.WithContext(ctx, log.NewConfig(log.LevelError))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/resources/myid/5/true?id=me&num=13&flag=true", strings.NewReader(`{"myString": "hello", "myInt": 5, "shouldFail": true}`))
	assert.Nil(t, err)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add(httpheader.Authorization, "valid-token")

	requester, err := endpoint.NewHttpRequester("/resources/{id}/{num}/{flag}", req)
	assert.Nil(t, err)

	var httpStatus int
	var responseBodyBytes []byte

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		httpStatus, responseBodyBytes = testEndpoint.Execute(ctx, requester)
	}()
	wg.Wait()

	if httpstatus.UnprocessableEntity != httpStatus {
		t.Errorf("expected HTTP status code to be %v, but got %v", httpstatus.OK, httpStatus)
	}

	var responseBody errorResponse
	if err := json.Unmarshal(responseBodyBytes, &responseBody); err != nil {
		t.Errorf("failed to unmarshal test response: %v", err)
	}

	if responseBody.Message != "[APP1] invalid input" {
		t.Errorf("expected 'message' to be '%v', but got '%v'", "[APP1] invalid input", responseBody.Message)
	}
}

func Test_Endpoint_Default_Error(t *testing.T) {
	t.Parallel()

	testEndpoint := createDefaultTestEndpoint()

	ctx := context.Local()
	ctx = log.WithContext(ctx, log.NewConfig(log.LevelError))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/resources/myid/5/true?id=me&num=13&flag=true", strings.NewReader(`{"myString": "hello", "myInt": 5, "shouldFail": true}`))
	assert.Nil(t, err)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add(httpheader.Authorization, "valid-token")

	requester, err := endpoint.NewHttpRequester("/resources/{id}/{num}/{flag}", req)
	assert.Nil(t, err)

	var httpStatus int
	var responseBodyBytes []byte

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		httpStatus, responseBodyBytes = testEndpoint.Execute(ctx, requester)
	}()
	wg.Wait()

	if httpstatus.UnprocessableEntity != httpStatus {
		t.Errorf("expected HTTP status code to be %v, but got %v", httpstatus.OK, httpStatus)
	}

	if string(responseBodyBytes) != `{"message":"[APP1] invalid input"}` {
		t.Errorf("expected 'message' to be '%v', but got '%v'", "[APP1] invalid input", string(responseBodyBytes))
	}
}

func Test_Endpoint_Auth_Error(t *testing.T) {
	t.Parallel()

	testEndpoint := createDefaultTestEndpoint()

	ctx := context.Local()
	ctx = log.WithContext(ctx, log.NewConfig(log.LevelError))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/resources/myid/5/true?id=me&num=13&flag=true", strings.NewReader(`{"myString": "hello", "myInt": 5, "shouldFail": true}`))
	assert.Nil(t, err)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	requester, err := endpoint.NewHttpRequester("/resources/{id}/{num}/{flag}", req)
	assert.Nil(t, err)

	var httpStatus int
	var responseBodyBytes []byte

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		httpStatus, responseBodyBytes = testEndpoint.Execute(ctx, requester)
	}()
	wg.Wait()

	if httpstatus.Unauthorized != httpStatus {
		t.Errorf("expected HTTP status code to be %v, but got %v", httpstatus.OK, httpStatus)
	}

	if string(responseBodyBytes) != `{"message":"[auth] invalid auth token"}` {
		t.Errorf("expected 'message' to be '%v', but got '%v'", "[auth] invalid auth token", string(responseBodyBytes))
	}
}
