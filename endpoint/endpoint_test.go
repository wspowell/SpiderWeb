package endpoint

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/wspowell/spiderweb/httpheader"
	"github.com/wspowell/spiderweb/httpstatus"

	"github.com/stretchr/testify/assert"
	"github.com/wspowell/context"
	"github.com/wspowell/errors"
	"github.com/wspowell/log"
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

func (self *user) Authorization(ctx context.Context, PeekHeader func(key string) []byte) (int, error) {
	var statusCode int

	if !bytes.EqualFold([]byte("valid-token"), PeekHeader(httpheader.Authorization)) {
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
	MyString   string `json:"my_string"`
	MyInt      int    `json:"my_int"`
	ShouldFail bool   `json:"fail"`
}

type myResponseBodyModel struct {
	MyString string `json:"output_string"`
	MyInt    int    `json:"output_int"`
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
		MyString: self.RequestBody.MyString,
		MyInt:    self.RequestBody.MyInt,
	}

	return httpstatus.OK, nil
}

func createTestEndpoint() *Endpoint {
	ctx := context.Local()
	log.WithContext(ctx, log.NewConfig(log.LevelError))

	dbClient := myDbClient{
		conn: "myconnection",
	}

	config := &Config{
		LogConfig:         log.NewConfig(log.LevelError),
		ErrorHandler:      myErrorHandler{},
		RequestValidator:  myRequestValidator{},
		ResponseValidator: myResponseValidator{},
		MimeTypeHandlers: map[string]*MimeTypeHandler{
			"application/json": jsonHandler(),
		},
		Resources: map[string]interface{}{
			"db": &dbClient,
		},
	}

	return NewEndpoint(ctx, config, &myEndpoint{})
}

func createDefaultTestEndpoint() *Endpoint {
	ctx := context.Local()
	log.WithContext(ctx, log.NewConfig(log.LevelError))

	dbClient := myDbClient{
		conn: "myconnection",
	}

	config := &Config{
		Resources: map[string]interface{}{
			"db": &dbClient,
		},
	}

	return NewEndpoint(ctx, config, &myEndpoint{})
}

func Test_Endpoint_Success(t *testing.T) {
	t.Parallel()

	endpoint := createTestEndpoint()

	req, err := http.NewRequest(http.MethodPost, "/resources/myid/5/true?id=me&num=13&flag=true", strings.NewReader(`{"my_string": "hello", "my_int": 5}`))
	assert.Nil(t, err)

	req.Header.Add(httpheader.ContentType, "application/json")
	req.Header.Add(httpheader.Accept, "application/json")
	req.Header.Add(httpheader.Authorization, "valid-token")

	requester := NewHttpRequester("/resources/{id}/{num}/{flag}", req)

	ctx := context.Local()
	ctx = log.WithContext(ctx, log.NewConfig(log.LevelError))
	var httpStatus int
	var responseBodyBytes []byte

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		httpStatus, responseBodyBytes = endpoint.Execute(ctx, requester)
	}()
	wg.Wait()

	if httpstatus.OK != httpStatus {
		t.Errorf("expected HTTP status code to be %v, but got %v", httpstatus.OK, httpStatus)
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

	req.Header.Add(httpheader.ContentType, "application/json")
	req.Header.Add(httpheader.Accept, "application/json")
	req.Header.Add(httpheader.Authorization, "valid-token")

	requester := NewHttpRequester("/resources/{id}/{num}/{flag}", req)

	ctx := context.Local()
	ctx = log.WithContext(ctx, log.NewConfig(log.LevelError))
	var httpStatus int
	var responseBodyBytes []byte

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		httpStatus, responseBodyBytes = endpoint.Execute(ctx, requester)
	}()
	wg.Wait()

	if httpstatus.OK != httpStatus {
		t.Errorf("expected HTTP status code to be %v, but got %v", httpstatus.OK, httpStatus)
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
	req.Header.Add(httpheader.Authorization, "valid-token")

	requester := NewHttpRequester("/resources/{id}/{num}/{flag}", req)

	ctx := context.Local()
	ctx = log.WithContext(ctx, log.NewConfig(log.LevelError))
	var httpStatus int
	var responseBodyBytes []byte

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		httpStatus, responseBodyBytes = endpoint.Execute(ctx, requester)
	}()
	wg.Wait()

	if httpstatus.UnprocessableEntity != httpStatus {
		t.Errorf("expected HTTP status code to be %v, but got %v", httpstatus.OK, httpStatus)
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
	req.Header.Add(httpheader.Authorization, "valid-token")

	requester := NewHttpRequester("/resources/{id}/{num}/{flag}", req)

	ctx := context.Local()
	ctx = log.WithContext(ctx, log.NewConfig(log.LevelError))
	var httpStatus int
	var responseBodyBytes []byte

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		httpStatus, responseBodyBytes = endpoint.Execute(ctx, requester)
	}()
	wg.Wait()

	if httpstatus.UnprocessableEntity != httpStatus {
		t.Errorf("expected HTTP status code to be %v, but got %v", httpstatus.OK, httpStatus)
	}

	if `{"message":"[APP1] invalid input"}` != string(responseBodyBytes) {
		t.Errorf("expected 'message' to be '%v', but got '%v'", "[APP1] invalid input", string(responseBodyBytes))
	}
}

func Test_Endpoint_Auth_Error(t *testing.T) {
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
		httpStatus, responseBodyBytes = endpoint.Execute(ctx, requester)
	}()
	wg.Wait()

	if httpstatus.Unauthorized != httpStatus {
		t.Errorf("expected HTTP status code to be %v, but got %v", httpstatus.OK, httpStatus)
	}

	if `{"message":"[auth] invalid auth token"}` != string(responseBodyBytes) {
		t.Errorf("expected 'message' to be '%v', but got '%v'", "[auth] invalid auth token", string(responseBodyBytes))
	}
}
