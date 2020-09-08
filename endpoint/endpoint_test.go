package endpoint

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"spiderweb/errors"
	"spiderweb/logging"
)

type errorResponse struct {
	InternalCode string `json:"internal_code"`
	Message      string `json:"message"`
}

type myErrorHandler struct{}

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

func (self myAuther) Auth(request *http.Request) (int, error) {
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
	Test         string
	RequestBody  *myRequestBodyModel  `spiderweb:"request,mime=test,validate"`
	ResponseBody *myResponseBodyModel `spiderweb:"response,mime=json,validate"`
}

func (self *myEndpoint) Handle(ctx *Context) (int, error) {
	ctx.Debug("handling myEndpoint")

	if self.RequestBody.ShouldFail {
		return http.StatusUnprocessableEntity, errors.New("APP1234", "invalid input")
	}

	self.ResponseBody = &myResponseBodyModel{
		MyString: self.RequestBody.MyString,
		MyInt:    self.RequestBody.MyInt,
	}

	return http.StatusOK, nil
}

func createTestEndpoint() *Endpoint {
	config := &Config{
		ErrorHandler:      myErrorHandler{},
		Auther:            myAuther{},
		RequestValidator:  myRequestValidator{},
		ResponseValidator: myResponseValidator{},
		MimeTypeHandlers: map[string]MimeTypeHandler{
			"test": MimeTypeHandler{
				Marshal: func(v interface{}) ([]byte, error) {
					return json.Marshal(v)
				},
				Unmarshal: func(data []byte, v interface{}) error {
					return json.Unmarshal(data, v)
				},
			},
		},
	}
	return NewEndpoint(config, &myEndpoint{})
}

func Test_Endpoint_Default_Success(t *testing.T) {
	logConfig := logging.NewConfig(logging.LevelDebug, map[string]interface{}{})

	req := httptest.NewRequest(http.MethodGet, "/", bytes.NewBuffer([]byte(`{"my_string": "hello", "my_int": 5}`)))

	ctx := NewContext(req, logging.NewLogger(logConfig))

	endpoint := createTestEndpoint()
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
	logConfig := logging.NewConfig(logging.LevelDebug, map[string]interface{}{})

	req := httptest.NewRequest(http.MethodGet, "/", bytes.NewBuffer([]byte(`{"my_string": "hello", "my_int": 5, "fail": true}`)))

	ctx := NewContext(req, logging.NewLogger(logConfig))

	endpoint := createTestEndpoint()
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
