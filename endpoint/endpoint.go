package endpoint

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"spiderweb/errors"
)

const (
	structTagKey = "spiderweb"

	structTagValueRequest  = "request"
	structTagValueResponse = "response"

	structTagOptionJson     = "json"
	structTagOptionValidate = "validate"
)

type Extensioner interface {
	Extension(ctx *Context, request *http.Request) (int, []byte)
}

type Endpoint struct {
	ErrorHandler      ErrorHandler
	Auther            Auther
	RequestValidator  RequestValidator
	ResponseValidator ResponseValidator
	handlerData       handlerTypeData
}

func NewEndpoint(handler Handler) *Endpoint {
	handlerData := newHandlerTypeData(handler)
	return &Endpoint{
		ErrorHandler:      ErrorResponse{},
		Auther:            Auth{},
		RequestValidator:  RequestValidation{},
		ResponseValidator: ResponseValidation{},
		handlerData:       handlerData,
	}
}

func (self *Endpoint) Execute(ctx *Context) (httpStatus int, responseBody []byte) {
	defer func() {
		if err := recover(); err != nil {
			panicErr := errors.New(ErrorCodePanic, "panic: %+v", err)
			httpStatus, responseBody = self.ErrorHandler.HandleError(ctx, http.StatusInternalServerError, panicErr)
		}
	}()

	handler := self.handlerData.newHandler().(Handler)

	requestBodyBytes, err := getRequestBody(ctx.Request())
	if err != nil {
		return self.ErrorHandler.HandleError(ctx, http.StatusInternalServerError, err)
	}

	if self.handlerData.shouldValidateRequest {
		if httpStatus, err := self.RequestValidator.ValidateRequest(ctx, requestBodyBytes); err != nil {
			validationError := errors.Wrap(ErrorCodeValidationError, err)
			return self.ErrorHandler.HandleError(ctx, httpStatus, validationError)
		}
	}

	if err := self.setHandlerRequestBody(handler, requestBodyBytes); err != nil {
		return self.ErrorHandler.HandleError(ctx, http.StatusInternalServerError, err)
	}

	httpStatus, err = handler.Handle(ctx)
	if err != nil {
		return self.ErrorHandler.HandleError(ctx, httpStatus, err)
	}

	responseBodyBytes, err := self.getHandlerResponseBody(handler)
	if err != nil {
		return self.ErrorHandler.HandleError(ctx, http.StatusInternalServerError, err)
	}

	if self.handlerData.shouldValidateResponse {
		if httpStatus, err := self.ResponseValidator.ValidateResponse(ctx, httpStatus, requestBodyBytes); err != nil {
			return self.ErrorHandler.HandleError(ctx, httpStatus, err)
		}
	}

	return httpStatus, responseBodyBytes
}

func getRequestBody(request *http.Request) ([]byte, error) {
	buffer := bytes.NewBuffer([]byte{})
	if _, err := io.Copy(buffer, request.Body); err != nil {
		return nil, errors.Wrap(ErrorCodeRequestBodyCopyFailure, err)
	}

	return buffer.Bytes(), nil
}

func (self *Endpoint) setHandlerRequestBody(handler Handler, requestBodyBytes []byte) error {
	if self.handlerData.requestBodyValue.IsValid() {

		handlerValue := reflect.ValueOf(handler)
		if self.handlerData.isStructPtr {
			handlerValue = handlerValue.Elem()
		}
		var requestValue reflect.Value
		if self.handlerData.isRequestPtr {
			requestValue = handlerValue.Field(self.handlerData.requestFieldNum)
		} else {
			requestValue = handlerValue.Field(self.handlerData.requestFieldNum)
		}
		var value interface{}
		if self.handlerData.isRequestPtr {
			value = requestValue.Addr().Interface()
		} else {
			value = requestValue.Addr().Interface()
		}

		if self.handlerData.isResponseJson {
			err := json.Unmarshal(requestBodyBytes, &value)
			if err != nil {
				return errors.Wrap(ErrorCodeRequestBodyJsonUnmarshalFailure, err)
			}
		}
	}

	return nil
}

func (self *Endpoint) getHandlerResponseBody(handler Handler) ([]byte, error) {
	handlerValue := reflect.ValueOf(handler)
	if self.handlerData.isStructPtr {
		handlerValue = handlerValue.Elem()
	}
	var responseValue reflect.Value
	if self.handlerData.isResponsePtr {
		responseValue = handlerValue.Field(self.handlerData.responseFieldNum)
	} else {
		responseValue = handlerValue.Field(self.handlerData.responseFieldNum)
	}
	var value interface{}
	if self.handlerData.isResponsePtr {
		value = responseValue.Elem().Interface()
	} else {
		value = responseValue.Interface()
	}

	if value != nil {
		if self.handlerData.isResponseJson {
			responseBodyBytes, err := json.Marshal(value)
			if err != nil {
				return nil, errors.Wrap(ErrorCodeResponseBodyJsonMarshalFailure, err)
			}
			if len(responseBodyBytes) == 4 && string(responseBodyBytes) == "null" {
				return nil, errors.New(ErrorCodeResponseBodyNull, "response body is null")
			}

			return responseBodyBytes, nil
		} else {
			// Default just dumps the value as a string.
			return []byte(fmt.Sprintf("%+v", value)), nil
		}
	}

	return nil, errors.New(ErrorCodeMissingResponseBody, "missing response body")
}

// handlerTypeData cached so that reflection is optimized.
type handlerTypeData struct {
	structValue       reflect.Value
	requestBodyValue  reflect.Value
	responseBodyValue reflect.Value

	isStructPtr   bool
	isRequestPtr  bool
	isResponsePtr bool

	requestFieldNum  int
	responseFieldNum int

	shouldValidateRequest  bool
	shouldValidateResponse bool

	isRequestJson  bool
	isResponseJson bool
}

func newHandlerTypeData(handler interface{}) handlerTypeData {
	var structValue reflect.Value
	var requestBodyValue reflect.Value
	var responseBodyValue reflect.Value
	var isStructPtr bool
	var isRequestPtr bool
	var isResponsePtr bool
	var requestFieldNum int
	var responseFieldNum int
	var shouldValidateRequest bool
	var shouldValidateResponse bool
	var isRequestJson bool
	var isResponseJson bool

	structValue = reflect.ValueOf(handler)
	if structValue.Kind() == reflect.Ptr {
		structValue = structValue.Elem()
		isStructPtr = true
	} else {
		panic("handler must be a reference")
	}

	for i := 0; i < structValue.NumField(); i++ {
		structFieldValue := structValue.Field(i)
		structField := structValue.Type().Field(i)
		if tagValue, exists := structField.Tag.Lookup(structTagKey); exists {
			tagValueParts := strings.Split(tagValue, ",")
			switch tagValueParts[0] {
			case structTagValueRequest:
				requestBodyValue = getFieldValue(structFieldValue)
				requestFieldNum = i
				isRequestPtr = structFieldValue.Kind() == reflect.Ptr
				shouldValidateRequest = hasStructTagOption(tagValue, structTagOptionValidate)
				isRequestJson = hasStructTagOption(tagValue, structTagOptionJson)
			case structTagValueResponse:
				responseBodyValue = getFieldValue(structFieldValue)
				responseFieldNum = i
				isResponsePtr = structFieldValue.Kind() == reflect.Ptr
				shouldValidateResponse = hasStructTagOption(tagValue, structTagOptionValidate)
				isResponseJson = hasStructTagOption(tagValue, structTagOptionJson)
			}
		}
	}

	return handlerTypeData{
		structValue:            structValue,
		requestBodyValue:       requestBodyValue,
		responseBodyValue:      responseBodyValue,
		isStructPtr:            isStructPtr,
		isRequestPtr:           isRequestPtr,
		isResponsePtr:          isResponsePtr,
		requestFieldNum:        requestFieldNum,
		responseFieldNum:       responseFieldNum,
		shouldValidateRequest:  shouldValidateRequest,
		shouldValidateResponse: shouldValidateResponse,
		isRequestJson:          isRequestJson,
		isResponseJson:         isResponseJson,
	}

}

func (self handlerTypeData) newHandler() interface{} {
	if !self.isStructPtr && !self.isRequestPtr && !self.isResponsePtr {
		handlerCopy := self.structValue.Interface()
		return handlerCopy
	}

	var handlerCopy reflect.Value
	if self.isStructPtr {
		handlerCopy = reflect.New(self.structValue.Type())
	} else {
		handlerCopy = reflect.ValueOf(self.structValue)
	}

	setNewFieldValue(handlerCopy, self.requestFieldNum, self.requestBodyValue)
	setNewFieldValue(handlerCopy, self.responseFieldNum, self.responseBodyValue)

	return handlerCopy.Interface()
}

func setNewFieldValue(structValue reflect.Value, fieldNum int, value reflect.Value) {
	// Only if the value is a pointer. Values will be zero initialized automatically.
	if value.Kind() == reflect.Ptr {
		requestBodyValue := structValue.Elem().Field(fieldNum)
		requestBodyValue.Set(reflect.New(value.Elem().Type()))
	}
}

func getFieldValue(structValue reflect.Value) reflect.Value {
	if structValue.IsZero() && structValue.Kind() == reflect.Ptr {
		newStructFieldValue := reflect.New(structValue.Type().Elem())
		structValue.Set(newStructFieldValue)
	}
	return structValue
}

func hasStructTagOption(structTag string, tagOption string) bool {
	for _, tagPart := range strings.Split(structTag, ",") {
		if tagPart == tagOption {
			return true
		}
	}

	return false
}
