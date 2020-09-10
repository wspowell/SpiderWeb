package endpoint

import (
	"fmt"
	"net/http"
	"time"

	"spiderweb/errors"
	"spiderweb/logging"
)

const (
	structTagKey = "spiderweb"

	structTagValueRequest  = "request"
	structTagValueResponse = "response"

	structTagOptionValidate = "validate"
)

type Config struct {
	LogConfig         logging.Configurer
	ErrorHandler      ErrorHandler
	Auther            Auther
	RequestValidator  RequestValidator
	ResponseValidator ResponseValidator
	MimeTypeHandlers  map[string]MimeTypeHandler
	Resources         map[string]ResourceFunc
	Timeout           time.Duration
}

// Clone the Config.
// This is necessary because MimeTypeHandlers is a map and there a reference.
func (self Config) Clone() Config {
	return Config{
		LogConfig:         self.LogConfig.Clone(),
		ErrorHandler:      self.ErrorHandler,
		Auther:            self.Auther,
		RequestValidator:  self.RequestValidator,
		ResponseValidator: self.ResponseValidator,
		MimeTypeHandlers:  self.copyMimeTypeHandlers(),
		Resources:         self.copyResourceFuncs(),
		Timeout:           self.Timeout,
	}
}

// copyMimeTypeHandlers to get a new instance of the map. This solves issues where
// other objects might mistakenly alter the original map through its copied reference.
func (self Config) copyMimeTypeHandlers() map[string]MimeTypeHandler {
	copy := make(map[string]MimeTypeHandler, len(self.MimeTypeHandlers))

	for mimeType, handler := range self.MimeTypeHandlers {
		copy[mimeType] = handler
	}

	return copy
}

func (self Config) copyResourceFuncs() map[string]ResourceFunc {
	copy := make(map[string]ResourceFunc, len(self.Resources))

	for resourceType, fn := range self.Resources {
		copy[resourceType] = fn
	}

	return copy
}

type Endpoint struct {
	Config      Config
	handlerData handlerTypeData
}

func NewEndpoint(config Config, handler Handler) *Endpoint {
	registerKnownMimeTypes(config.MimeTypeHandlers)

	handlerData := newHandlerTypeData(handler)

	return &Endpoint{
		Config:      config,
		handlerData: handlerData,
	}
}

func (self *Endpoint) Execute(ctx *Context) (httpStatus int, responseBody []byte) {
	defer func() {
		if err := recover(); err != nil {
			ctx.Error("panic: %+v", errors.New("ERROR", fmt.Sprintf("%+v", err)))
			httpStatus, responseBody = self.Config.ErrorHandler.HandleError(ctx, http.StatusInternalServerError, ErrorPanic)
		}
	}()

	var err error

	if !ctx.ShouldContinue() {
		return self.Config.ErrorHandler.HandleError(ctx, http.StatusRequestTimeout, ErrorRequestTimeout)
	}

	if httpStatus, err = self.Config.Auther.Auth(ctx.Request()); err != nil {
		return self.Config.ErrorHandler.HandleError(ctx, httpStatus, err)
	}

	handlerAlloc := self.handlerData.allocateHandler(self.Config.Resources)

	// Handle Request
	{
		if !ctx.ShouldContinue() {
			return self.Config.ErrorHandler.HandleError(ctx, http.StatusRequestTimeout, ErrorRequestTimeout)
		}

		requestBodyBytes := ctx.Request().Body()

		if self.handlerData.shouldValidateRequest {
			if httpStatus, validationFailure := self.Config.RequestValidator.ValidateRequest(ctx, requestBodyBytes); err != nil {
				// Validation failures are not hard errors and should be passed through to the error handler.
				// The failure is passed through since it is assumed this error contains information to be returned in the response.
				return self.Config.ErrorHandler.HandleError(ctx, httpStatus, validationFailure)
			}
		}

		if err := self.setHandlerRequestBody(ctx, handlerAlloc.requestBody, requestBodyBytes); err != nil {
			return self.Config.ErrorHandler.HandleError(ctx, http.StatusInternalServerError, err)
		}
	}

	if !ctx.ShouldContinue() {
		return self.Config.ErrorHandler.HandleError(ctx, http.StatusRequestTimeout, ErrorRequestTimeout)
	}

	// Run the endpoint handler.
	if httpStatus, err = handlerAlloc.handler.Handle(ctx); err != nil {
		return self.Config.ErrorHandler.HandleError(ctx, httpStatus, err)
	}

	// Handle Response
	{
		if !ctx.ShouldContinue() {
			return self.Config.ErrorHandler.HandleError(ctx, http.StatusRequestTimeout, ErrorRequestTimeout)
		}

		if responseBody, err = self.getHandlerResponseBody(ctx, handlerAlloc.responseBody); err != nil {
			return self.Config.ErrorHandler.HandleError(ctx, http.StatusInternalServerError, err)
		}

		if self.handlerData.shouldValidateResponse {
			if httpStatus, validationFailure := self.Config.ResponseValidator.ValidateResponse(ctx, httpStatus, responseBody); err != nil {
				// Validation failures are not hard errors and should be passed through to the error handler.
				// The failure is passed through since it is assumed this error contains information to be returned in the response.
				return self.Config.ErrorHandler.HandleError(ctx, httpStatus, validationFailure)
			}
		}
	}

	return httpStatus, responseBody
}

func (self *Endpoint) setHandlerRequestBody(ctx *Context, requestBody interface{}, requestBodyBytes []byte) error {
	if requestBody != nil {
		if mimeHandler, exists := self.Config.MimeTypeHandlers[self.handlerData.requestMimeType]; exists {
			ctx.requestCtx.SetContentType(mimeHandler.MimeType)
			err := mimeHandler.Unmarshal(requestBodyBytes, &requestBody)
			if err != nil {
				ctx.Error("failed to unmarshal request body: %v", err)
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
		if mimeHandler, exists := self.Config.MimeTypeHandlers[self.handlerData.responseMimeType]; exists {
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
