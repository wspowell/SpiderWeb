package endpointtest

import (
	"encoding/json"
	"net/http"

	"spiderweb/endpoint"
	"spiderweb/errors"
	"spiderweb/logging"

	"github.com/valyala/fasthttp"
)

func createTestEndpoint() *endpoint.Endpoint {
	config := endpoint.Config{
		LogConfig:         logging.NewConfig(logging.LevelDebug, map[string]interface{}{}),
		ErrorHandler:      myErrorHandler{},
		Auther:            myAuther{},
		RequestValidator:  myRequestValidator{},
		ResponseValidator: myResponseValidator{},
		MimeTypeHandlers:  map[string]endpoint.MimeTypeHandler{},
		Resources:         map[string]endpoint.ResourceFunc{},
	}

	return endpoint.NewEndpoint(config.Clone(), &myEndpoint{})
}

type errorResponse struct {
	InternalCode string `json:"internal_code"`
	Message      string `json:"message"`
}

type myErrorHandler struct{}

func (self myErrorHandler) HandleError(ctx *endpoint.Context, httpStatus int, err error) (int, []byte) {
	if endpoint.HasFrameworkError(err) {
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

func (self myAuther) Auth(request *fasthttp.Request) (int, error) {
	return http.StatusOK, nil
}

type myRequestValidator struct{}

func (self myRequestValidator) ValidateRequest(ctx *endpoint.Context, requestBodyBytes []byte) (int, error) {
	return http.StatusOK, nil
}

type myResponseValidator struct{}

func (self myResponseValidator) ValidateResponse(ctx *endpoint.Context, httpStatus int, responseBodyBytes []byte) (int, error) {
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
	RequestBody  *myRequestBodyModel  `spiderweb:"request,mime=json,validate"`
	ResponseBody *myResponseBodyModel `spiderweb:"response,mime=json,validate"`
}

func (self *myEndpoint) Handle(ctx *endpoint.Context) (int, error) {
	ctx.Debug("handling myEndpoint")

	if self.RequestBody.ShouldFail {
		return http.StatusUnprocessableEntity, errors.New("APP1234", "invalid input")
	}

	self.ResponseBody = &myResponseBodyModel{
		MyString: self.RequestBody.MyString,
		MyInt:    self.RequestBody.MyInt,
	}

	return http.StatusCreated, nil
}
