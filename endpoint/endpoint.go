package endpoint

import (
	"bytes"
	"io"
	"net/http"
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
			ctx.Error("panic: %+v", err)
			httpStatus, responseBody = self.config.ErrorHandler.HandleError(ctx, http.StatusInternalServerError, ErrorPanic)
		}
	}()

	var err error
	if httpStatus, err = self.config.Auther.Auth(ctx.Request()); err != nil {
		return self.config.ErrorHandler.HandleError(ctx, httpStatus, err)
	}

	handler := self.handlerData.newHandler().(Handler)

	// Handle Request
	{
		var requestBodyBytes []byte
		if requestBodyBytes, err = getRequestBody(ctx); err != nil {
			return self.config.ErrorHandler.HandleError(ctx, http.StatusInternalServerError, err)
		}

		if self.handlerData.shouldValidateRequest {
			if httpStatus, validationFailure := self.config.RequestValidator.ValidateRequest(ctx, requestBodyBytes); err != nil {
				// Validation failures are not hard errors and should be passed through to the error handler.
				// The failure is passed through since it is assumed this error contains information to be returned in the response.
				return self.config.ErrorHandler.HandleError(ctx, httpStatus, validationFailure)
			}
		}

		if err := self.setHandlerRequestBody(ctx, handler, requestBodyBytes); err != nil {
			return self.config.ErrorHandler.HandleError(ctx, http.StatusInternalServerError, err)
		}
	}

	// Run the endpoint handler.
	if httpStatus, err = handler.Handle(ctx); err != nil {
		return self.config.ErrorHandler.HandleError(ctx, httpStatus, err)
	}

	// Handle Response
	{
		if responseBody, err = self.getHandlerResponseBody(ctx, handler); err != nil {
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

func (self *Endpoint) setHandlerRequestBody(ctx *Context, handler Handler, requestBodyBytes []byte) error {
	value := newStructFromHandler(handler, self.handlerData.isStructPtr, self.handlerData.requestFieldNum)
	if value != nil {
		if mimeHandler, exists := self.config.MimeTypeHandlers[self.handlerData.requestMimeType]; exists {
			err := mimeHandler.Unmarshal(requestBodyBytes, &value)
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

func (self *Endpoint) getHandlerResponseBody(ctx *Context, handler Handler) ([]byte, error) {
	value := newStructFromHandler(handler, self.handlerData.isStructPtr, self.handlerData.responseFieldNum)
	if value != nil {
		if mimeHandler, exists := self.config.MimeTypeHandlers[self.handlerData.responseMimeType]; exists {
			responseBodyBytes, err := mimeHandler.Marshal(value)
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

func getRequestBody(ctx *Context) ([]byte, error) {
	request := ctx.Request()

	var bodyBytes []byte
	if request.ContentLength > 0 {
		// Take advantage of the request content length, if available.
		bodyBytes = make([]byte, 0, request.ContentLength)
	}

	buffer := bytes.NewBuffer(bodyBytes)
	if _, err := io.Copy(buffer, request.Body); err != nil {
		ctx.Error("failed to read request: %v", err)
		return nil, ErrorRequestBodyReadFailure
	}

	return buffer.Bytes(), nil
}
