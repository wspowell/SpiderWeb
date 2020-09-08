package endpoint

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"spiderweb/errors"
)

const (
	structTagKey = "spiderweb"

	structTagValueRequest  = "request"
	structTagValueResponse = "response"

	structTagOptionValidate = "validate"
)

type Config struct {
	ErrorHandler      ErrorHandler
	Auther            Auther
	RequestValidator  RequestValidator
	ResponseValidator ResponseValidator
	MimeTypeHandlers  map[string]MimeTypeHandler
}

type Endpoint struct {
	config      *Config
	handlerData handlerTypeData
}

func NewEndpoint(config *Config, handler Handler) *Endpoint {

	registerKnownMimeTypes(config.MimeTypeHandlers)

	handlerData := newHandlerTypeData(handler)

	return &Endpoint{
		config:      config,
		handlerData: handlerData,
	}
}

func (self *Endpoint) Execute(ctx *Context) (httpStatus int, responseBody []byte) {
	defer func() {
		if err := recover(); err != nil {
			ctx.Error("panic: %+v", errors.New("ERROR", fmt.Sprintf("%+v", err)))
			httpStatus, responseBody = self.config.ErrorHandler.HandleError(ctx, http.StatusInternalServerError, ErrorPanic)
		}
	}()

	var err error
	if httpStatus, err = self.config.Auther.Auth(ctx.Request()); err != nil {
		return self.config.ErrorHandler.HandleError(ctx, httpStatus, err)
	}

	handlerAlloc := self.handlerData.allocateHandler()

	// Handle Request
	{
		requestBodyBytes, err := readRequestBody(ctx)
		if err != nil {
			return self.config.ErrorHandler.HandleError(ctx, http.StatusInternalServerError, err)
		}

		if self.handlerData.shouldValidateRequest {
			if httpStatus, validationFailure := self.config.RequestValidator.ValidateRequest(ctx, requestBodyBytes); err != nil {
				// Validation failures are not hard errors and should be passed through to the error handler.
				// The failure is passed through since it is assumed this error contains information to be returned in the response.
				return self.config.ErrorHandler.HandleError(ctx, httpStatus, validationFailure)
			}
		}

		if err := self.setHandlerRequestBody(ctx, handlerAlloc.requestBody, requestBodyBytes); err != nil {
			return self.config.ErrorHandler.HandleError(ctx, http.StatusInternalServerError, err)
		}
	}

	// Run the endpoint handler.
	if httpStatus, err = handlerAlloc.handler.Handle(ctx); err != nil {
		return self.config.ErrorHandler.HandleError(ctx, httpStatus, err)
	}

	// Handle Response
	{
		if responseBody, err = self.getHandlerResponseBody(ctx, handlerAlloc.responseBody); err != nil {
			return self.config.ErrorHandler.HandleError(ctx, http.StatusInternalServerError, err)
		}

		if self.handlerData.shouldValidateResponse {
			if httpStatus, validationFailure := self.config.ResponseValidator.ValidateResponse(ctx, httpStatus, responseBody); err != nil {
				// Validation failures are not hard errors and should be passed through to the error handler.
				// The failure is passed through since it is assumed this error contains information to be returned in the response.
				return self.config.ErrorHandler.HandleError(ctx, httpStatus, validationFailure)
			}
		}
	}

	return httpStatus, responseBody
}

func (self *Endpoint) setHandlerRequestBody(ctx *Context, requestBody interface{}, requestBodyBytes []byte) error {
	if requestBody != nil {
		if mimeHandler, exists := self.config.MimeTypeHandlers[self.handlerData.requestMimeType]; exists {
			err := mimeHandler.Unmarshal(requestBodyBytes, &requestBody)
			if err != nil {
				return ErrorRequestBodyUnmarshalFailure
			}
			return nil
		}

		ctx.Error("unknown request mime type: %v", self.handlerData.requestMimeType)
		return ErrorRequestUnknownMimeType
	}
	return nil
}

func (self *Endpoint) getHandlerResponseBody(ctx *Context, responseBody interface{}) ([]byte, error) {
	if responseBody != nil {
		if mimeHandler, exists := self.config.MimeTypeHandlers[self.handlerData.responseMimeType]; exists {
			responseBodyBytes, err := mimeHandler.Marshal(responseBody)
			if err != nil {
				ctx.Error("failed to marshal response: %v", err)
				return nil, ErrorResponseBodyMarshalFailure
			}
			if len(responseBodyBytes) == 4 && string(responseBodyBytes) == "null" {
				return nil, ErrorResponseBodyNull
			}
			return responseBodyBytes, nil
		}
		ctx.Error("unknown response mime type: %v", self.handlerData.responseMimeType)
		return nil, ErrorResponseUnknownMimeType
	}

	return nil, ErrorResponseBodyMissing
}

// readRequestBody into a byte array.
func readRequestBody(ctx *Context) ([]byte, error) {
	request := ctx.Request()

	var bodyBytes []byte
	if request.ContentLength > 0 {
		// Take advantage of the request content length, if available.
		bodyBytes = make([]byte, 0, request.ContentLength)
	}

	// Note: This uses less memory and is more efficient than ioutil.ReadAll().
	buffer := bytes.NewBuffer(bodyBytes)
	if _, err := io.Copy(buffer, request.Body); err != nil {
		ctx.Error("failed to read request: %v", err)
		return nil, ErrorRequestBodyReadFailure
	}

	return buffer.Bytes(), nil
}
