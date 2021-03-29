package endpoint

import (
	"bytes"
	"net/http"
	"time"

	"github.com/wspowell/errors"
	"github.com/wspowell/logging"
	"github.com/wspowell/spiderweb/profiling"
)

const (
	structTagKey = "spiderweb"

	structTagValueRequest  = "request"
	structTagValueResponse = "response"

	structTagOptionValidate = "validate"
)

const (
	HeaderAccept = "Accept"
)

var (
	nullBytes = []byte("null")
)

// Config defines the behavior of an endpoint.
// Endpoint behavior is interface driven and can be completely modified by an application.
type Config struct {
	LogConfig         logging.Configer
	ErrorHandler      ErrorHandler
	Auther            Auther
	RequestValidator  RequestValidator
	ResponseValidator ResponseValidator
	MimeTypeHandlers  MimeTypeHandlers
	Resources         map[string]ResourceFunc
	Timeout           time.Duration
}

// Clone the Config.
// This is necessary because MimeTypeHandlers is a map and therefore a reference.
func (self Config) Clone() Config {
	return Config{
		LogConfig:         self.LogConfig,
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
func (self Config) copyMimeTypeHandlers() MimeTypeHandlers {
	copy := make(MimeTypeHandlers, len(self.MimeTypeHandlers))

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

// Endpoint defines the behavior of a given handler.
type Endpoint struct {
	Config      Config
	handlerData handlerTypeData
}

// Create a new endpoint that will run the given handler.
// This will be created by the Server during normal operations.
func NewEndpoint(config Config, handler Handler) *Endpoint {
	registerKnownMimeTypes(config.MimeTypeHandlers)

	handlerData := newHandlerTypeData(handler)

	return &Endpoint{
		Config:      config,
		handlerData: handlerData,
	}
}

// Execute the endpoint and run the endpoint handler.
func (self *Endpoint) Execute(ctx *Context) (httpStatus int, responseBody []byte) {
	defer func() {
		if err := recover(); err != nil {
			ctx.Error("panic: %+v", errors.New("ERROR", "%+v", err))
			ctx.requestCtx.SetContentType(self.Config.ErrorHandler.MimeType())
			httpStatus, responseBody = self.Config.ErrorHandler.HandleError(ctx, http.StatusInternalServerError, ErrorPanic)
		}
	}()

	defer profiling.Profile(ctx, ctx.HttpMethod+" "+ctx.MatchedPath).Finish()

	var err error

	// Content-Type and Accept
	var requestMimeType MimeTypeHandler
	var responseMimeType MimeTypeHandler
	{
		var ok bool

		if self.handlerData.hasRequest {
			contentType := ctx.Request().Header.ContentType()
			if len(contentType) == 0 {
				ctx.requestCtx.SetContentType(self.Config.ErrorHandler.MimeType())
				return self.Config.ErrorHandler.HandleError(ctx, http.StatusUnsupportedMediaType, errors.New(InternalCodeRequestMimeTypeMissing, "Content-Type MIME type not provided"))
			}

			requestMimeType, ok = self.Config.MimeTypeHandlers.Get(contentType, self.handlerData.requestMimeTypes)
			if !ok {
				ctx.requestCtx.SetContentType(self.Config.ErrorHandler.MimeType())
				return self.Config.ErrorHandler.HandleError(ctx, http.StatusUnsupportedMediaType, errors.New(InternalCodeRequestMimeTypeUnsupported, "Content-Type MIME type not supported: %s", contentType))
			}
		}

		if self.handlerData.hasResponse {
			accept := ctx.Request().Header.Peek(HeaderAccept)
			if len(accept) == 0 {
				ctx.requestCtx.SetContentType(self.Config.ErrorHandler.MimeType())
				return self.Config.ErrorHandler.HandleError(ctx, http.StatusUnsupportedMediaType, errors.New(InternalCodeResponseMimeTypeMissing, "Accept MIME type not provided"))
			}

			responseMimeType, ok = self.Config.MimeTypeHandlers.Get(accept, self.handlerData.responseMimeTypes)
			if !ok {
				ctx.requestCtx.SetContentType(requestMimeType.MimeType)
				return self.Config.ErrorHandler.HandleError(ctx, http.StatusUnsupportedMediaType, errors.New(InternalCodeResponseMimeTypeUnsupported, "Accept MIME type not supported: %s", accept))
			}
		}
	}

	if !ctx.ShouldContinue() {
		ctx.requestCtx.SetContentType(responseMimeType.MimeType)
		return self.Config.ErrorHandler.HandleError(ctx, http.StatusRequestTimeout, ErrorRequestTimeout)
	}

	// Authentication
	{
		authTimer := profiling.Profile(ctx, "Auth")

		httpStatus, err = self.Config.Auther.Auth(ctx, ctx.Request().Header.VisitAll)
		authTimer.Finish()
		if err != nil {
			ctx.requestCtx.SetContentType(responseMimeType.MimeType)
			return self.Config.ErrorHandler.HandleError(ctx, httpStatus, err)
		}
	}

	if !ctx.ShouldContinue() {
		ctx.requestCtx.SetContentType(responseMimeType.MimeType)
		return self.Config.ErrorHandler.HandleError(ctx, http.StatusRequestTimeout, ErrorRequestTimeout)
	}

	allocateTimer := profiling.Profile(ctx, "Allocate")
	handlerAlloc := self.handlerData.allocateHandler()

	self.handlerData.setResources(handlerAlloc.handlerValue, self.Config.Resources)
	self.handlerData.setPathParameters(handlerAlloc.handlerValue, ctx.requestCtx)
	self.handlerData.setQueryParameters(handlerAlloc.handlerValue, ctx.requestCtx)
	allocateTimer.Finish()

	// Handle Request
	{
		if !ctx.ShouldContinue() {
			ctx.requestCtx.SetContentType(responseMimeType.MimeType)
			return self.Config.ErrorHandler.HandleError(ctx, http.StatusRequestTimeout, ErrorRequestTimeout)
		}

		requestBodyBytes := ctx.Request().Body()

		if self.handlerData.shouldValidateRequest {
			validateTimer := profiling.Profile(ctx, "ValidateRequest")
			httpStatus, validationFailure := self.Config.RequestValidator.ValidateRequest(ctx, requestBodyBytes)
			validateTimer.Finish()
			if err != nil {
				// Validation failures are not hard errors and should be passed through to the error handler.
				// The failure is passed through since it is assumed this error contains information to be returned in the response.
				ctx.requestCtx.SetContentType(responseMimeType.MimeType)
				return self.Config.ErrorHandler.HandleError(ctx, httpStatus, validationFailure)
			}
		}

		populateRequestTimer := profiling.Profile(ctx, "UnmarshalRequest")

		err := self.setHandlerRequestBody(ctx, requestMimeType, handlerAlloc.requestBody, requestBodyBytes)
		populateRequestTimer.Finish()
		if err != nil {
			ctx.requestCtx.SetContentType(responseMimeType.MimeType)
			return self.Config.ErrorHandler.HandleError(ctx, http.StatusInternalServerError, err)
		}
	}

	if !ctx.ShouldContinue() {
		ctx.requestCtx.SetContentType(responseMimeType.MimeType)
		return self.Config.ErrorHandler.HandleError(ctx, http.StatusRequestTimeout, ErrorRequestTimeout)
	}

	// Run the endpoint handler.
	handleTimer := profiling.Profile(ctx, "Handle")
	httpStatus, err = handlerAlloc.handler.Handle(ctx)
	handleTimer.Finish()
	if err != nil {
		ctx.requestCtx.SetContentType(responseMimeType.MimeType)
		return self.Config.ErrorHandler.HandleError(ctx, httpStatus, err)
	}

	// Handle Response
	{
		if !ctx.ShouldContinue() {
			ctx.requestCtx.SetContentType(responseMimeType.MimeType)
			return self.Config.ErrorHandler.HandleError(ctx, http.StatusRequestTimeout, ErrorRequestTimeout)
		}

		populateResponseTimer := profiling.Profile(ctx, "MarshalResponseBody")
		responseBody, err = self.getHandlerResponseBody(ctx, responseMimeType, handlerAlloc.responseBody)
		populateResponseTimer.Finish()
		if err != nil {
			ctx.requestCtx.SetContentType(responseMimeType.MimeType)
			return self.Config.ErrorHandler.HandleError(ctx, http.StatusInternalServerError, err)
		}

		if self.handlerData.shouldValidateResponse {
			validateResponseTimer := profiling.Profile(ctx, "ValidateResponse")
			httpStatus, validationFailure := self.Config.ResponseValidator.ValidateResponse(ctx, httpStatus, responseBody)
			validateResponseTimer.Finish()
			if err != nil {
				// Validation failures are not hard errors and should be passed through to the error handler.
				// The failure is passed through since it is assumed this error contains information to be returned in the response.
				ctx.requestCtx.SetContentType(responseMimeType.MimeType)
				return self.Config.ErrorHandler.HandleError(ctx, httpStatus, validationFailure)
			}
		}
	}

	return httpStatus, responseBody
}

func (self *Endpoint) setHandlerRequestBody(ctx *Context, mimeHandler MimeTypeHandler, requestBody interface{}, requestBodyBytes []byte) error {
	if requestBody != nil {
		if err := mimeHandler.Unmarshal(requestBodyBytes, requestBody); err != nil {
			ctx.Error("failed to unmarshal request body: %v", err)
			return ErrorRequestBodyUnmarshalFailure
		}
	}
	return nil
}

func (self *Endpoint) getHandlerResponseBody(ctx *Context, mimeHandler MimeTypeHandler, responseBody interface{}) ([]byte, error) {
	if responseBody != nil {
		ctx.requestCtx.SetContentType(mimeHandler.MimeType)
		responseBodyBytes, err := mimeHandler.Marshal(responseBody)
		if err != nil {
			ctx.Error("failed to marshal response: %v", err)
			return nil, ErrorResponseBodyMarshalFailure
		}
		if len(responseBodyBytes) == 4 && bytes.Equal(responseBodyBytes, nullBytes) {
			return nil, ErrorResponseBodyNull
		}
		return responseBodyBytes, nil
	}

	// No response body.
	return nil, nil
}
