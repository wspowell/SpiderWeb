package endpoint

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/wspowell/errors"
	"github.com/wspowell/log"
	"github.com/wspowell/spiderweb/profiling"
)

const (
	structTagKey = "spiderweb"

	structTagValueRequest  = "request"
	structTagValueResponse = "response"

	structTagOptionValidate = "validate"
)

var (
	nullBytes = []byte("null")
)

// Config defines the behavior of an endpoint.
// Endpoint behavior is interface driven and can be completely modified by an application.
// The values in the config must never be modified by an endpoint.
type Config struct {
	LogConfig         log.Configer
	ErrorHandler      ErrorHandler
	Auther            Auther
	RequestValidator  RequestValidator
	ResponseValidator ResponseValidator
	MimeTypeHandlers  MimeTypeHandlers
	Resources         map[string]interface{}
	Timeout           time.Duration
}

// Endpoint defines the behavior of a given handler.
type Endpoint struct {
	Config *Config

	handlerData handlerTypeData
}

// Create a new endpoint that will run the given handler.
// This will be created by the Server during normal operations.
func NewEndpoint(config *Config, handler Handler) *Endpoint {
	configClone := &Config{}

	// Set defaults, if not set.

	if config.ErrorHandler == nil {
		configClone.ErrorHandler = defaultErrorHandler{}
	} else {
		configClone.ErrorHandler = config.ErrorHandler
	}

	if config.MimeTypeHandlers == nil {
		configClone.MimeTypeHandlers = NewMimeTypeHandlers()
	} else {
		configClone.MimeTypeHandlers = config.MimeTypeHandlers
	}

	if config.Resources == nil {
		configClone.Resources = map[string]interface{}{}
	} else {
		configClone.Resources = config.Resources
	}

	if config.Timeout == 0 {
		configClone.Timeout = 30 * time.Second
	} else {
		configClone.Timeout = config.Timeout
	}

	return &Endpoint{
		Config: configClone,

		handlerData: newHandlerTypeData(handler),
	}
}

func (self *Endpoint) Name() string {
	return self.handlerData.structName
}

// Execute the endpoint and run the endpoint handler.
func (self *Endpoint) Execute(ctx *Context) (httpStatus int, responseBody []byte) {
	var responseMimeType *MimeTypeHandler

	defer func() {
		if errPanic := recover(); errPanic != nil {
			log.Error(ctx, "panic: %+v", errors.New("ERROR", "%+v", errPanic))
			httpStatus, responseBody = self.processErrorResponse(ctx, responseMimeType, http.StatusInternalServerError, ErrorPanic)
		}
	}()

	defer profiling.Profile(ctx, string(ctx.requester.Method())+" "+ctx.requester.MatchedPath()).Finish()

	ctx.requester.SetResponseHeader("X-Request-Id", ctx.requester.RequestId())

	// Setup log.
	{
		// Every invocation of an endpoint is guaranteed to get its own logger instance.
		// See: log.WithContext()
		log.Tag(ctx, "request_id", ctx.requester.RequestId())
		log.Tag(ctx, "method", string(ctx.requester.Method()))
		log.Tag(ctx, "route", ctx.requester.MatchedPath())
		log.Tag(ctx, "path", string(ctx.requester.Path()))
		log.Tag(ctx, "action", self.Name())

		// Each path parameter is added as a log tag.
		// Note: It helps if the path parameter name is descriptive.
		for param := range self.handlerData.pathParameters {
			if value, ok := ctx.requester.PathParam(param); ok {
				log.Tag(ctx, param, value)
			}
		}
	}

	log.Trace(ctx, "executing endpoint")

	var err error

	// Content-Type and Accept
	var requestMimeType *MimeTypeHandler
	{
		var ok bool

		if self.handlerData.hasRequestBody {
			log.Trace(ctx, "processing request body mime type")

			contentType := ctx.requester.ContentType()
			if len(contentType) == 0 {
				log.Debug(ctx, "header Content-Type not found")
				return self.processErrorResponse(ctx, responseMimeType, http.StatusUnsupportedMediaType, errors.New(InternalCodeRequestMimeTypeMissing, "Content-Type MIME type not provided"))
			}

			requestMimeType, ok = self.Config.MimeTypeHandlers.Get(contentType, self.handlerData.requestMimeTypes)
			if !ok {
				log.Debug(ctx, "mime type handler not available: %s", contentType)
				return self.processErrorResponse(ctx, responseMimeType, http.StatusUnsupportedMediaType, errors.New(InternalCodeRequestMimeTypeUnsupported, "Content-Type MIME type not supported: %s", contentType))
			}

			log.Debug(ctx, "found request mime type handler: %s", contentType)
		}

		if self.handlerData.hasResponseBody {
			log.Trace(ctx, "processing response body mime type")

			accept := ctx.requester.Accept()
			if len(accept) == 0 {
				log.Debug(ctx, "header Accept not found")
				return self.processErrorResponse(ctx, responseMimeType, http.StatusUnsupportedMediaType, errors.New(InternalCodeResponseMimeTypeMissing, "Accept MIME type not provided"))
			}

			responseMimeType, ok = self.Config.MimeTypeHandlers.Get(accept, self.handlerData.responseMimeTypes)
			if !ok {
				log.Debug(ctx, "mime type handler not available: %s", accept)
				return self.processErrorResponse(ctx, responseMimeType, http.StatusUnsupportedMediaType, errors.New(InternalCodeResponseMimeTypeUnsupported, "Accept MIME type not supported: %s", accept))
			}
			// All responses after this must be marshalable to the mime type.
			ctx.requester.SetResponseContentType(responseMimeType.MimeType)

			log.Debug(ctx, "found response mime type handler: %s", accept)
		}
	}

	if !ctx.ShouldContinue() {
		log.Debug(ctx, "request canceled or timed out")
		return self.processErrorResponse(ctx, responseMimeType, http.StatusRequestTimeout, ErrorRequestTimeout)
	}

	// Authentication
	{
		authTimer := profiling.Profile(ctx, "Auth")

		if self.Config.Auther != nil {
			log.Trace(ctx, "processing auth handler")

			httpStatus, err = self.Config.Auther.Auth(ctx, ctx.requester.VisitHeaders)
			authTimer.Finish()
			if err != nil {
				log.Debug(ctx, "auth failed")
				return self.processErrorResponse(ctx, responseMimeType, httpStatus, err)
			}
		}
	}

	if !ctx.ShouldContinue() {
		log.Debug(ctx, "request canceled or timed out")
		return self.processErrorResponse(ctx, responseMimeType, http.StatusRequestTimeout, ErrorRequestTimeout)
	}

	log.Trace(ctx, "allocating handler")

	allocateTimer := profiling.Profile(ctx, "Allocate")
	handlerAlloc := self.handlerData.allocateHandler()

	self.handlerData.setResources(handlerAlloc.handlerValue, self.Config.Resources)
	self.handlerData.setPathParameters(handlerAlloc.handlerValue, ctx.requester)
	self.handlerData.setQueryParameters(handlerAlloc.handlerValue, ctx.requester)
	allocateTimer.Finish()

	// Handle Request
	{
		if !ctx.ShouldContinue() {
			log.Debug(ctx, "request canceled or timed out")
			return self.processErrorResponse(ctx, responseMimeType, http.StatusRequestTimeout, ErrorRequestTimeout)
		}

		if self.handlerData.hasRequestBody {
			log.Trace(ctx, "processing request body")

			requestBodyBytes := ctx.requester.RequestBody()

			populateRequestTimer := profiling.Profile(ctx, "UnmarshalRequest")
			err = self.setHandlerRequestBody(ctx, requestMimeType, handlerAlloc.requestBody, requestBodyBytes)
			populateRequestTimer.Finish()
			if err != nil {
				log.Debug(ctx, "failed processing request body")
				return self.processErrorResponse(ctx, responseMimeType, http.StatusBadRequest, err)
			}

			if self.Config.RequestValidator != nil && self.handlerData.shouldValidateRequest {
				log.Trace(ctx, "processing validation handler")

				validateTimer := profiling.Profile(ctx, "ValidateRequest")
				var validationFailure error
				httpStatus, validationFailure = self.Config.RequestValidator.ValidateRequest(ctx, requestBodyBytes)
				validateTimer.Finish()
				if validationFailure != nil {
					log.Debug(ctx, "failed request body validation")

					// Validation failures are not hard errors and should be passed through to the error handler.
					// The failure is passed through since it is assumed this error contains information to be returned in the response.
					return self.processErrorResponse(ctx, responseMimeType, httpStatus, validationFailure)
				}
			}
		}
	}

	if !ctx.ShouldContinue() {
		log.Debug(ctx, "request canceled or timed out")
		return self.processErrorResponse(ctx, responseMimeType, http.StatusRequestTimeout, ErrorRequestTimeout)
	}

	// Run the endpoint handler.
	log.Trace(ctx, "running endpoint handler")
	handleTimer := profiling.Profile(ctx, self.Name()+".Handle()")
	httpStatus, err = handlerAlloc.handler.Handle(ctx)
	handleTimer.Finish()
	if err != nil {
		log.Debug(ctx, "handler error")
		return self.processErrorResponse(ctx, responseMimeType, httpStatus, err)
	}

	// Handle Response
	{
		if !ctx.ShouldContinue() {
			log.Debug(ctx, "request canceled or timed out")
			return self.processErrorResponse(ctx, responseMimeType, http.StatusRequestTimeout, ErrorRequestTimeout)
		}

		populateResponseTimer := profiling.Profile(ctx, "MarshalResponseBody")
		responseBody, err = self.getHandlerResponseBody(ctx, responseMimeType, handlerAlloc.responseBody)
		populateResponseTimer.Finish()
		if err != nil {
			log.Debug(ctx, "failed processing response")
			return self.processErrorResponse(ctx, responseMimeType, http.StatusInternalServerError, err)
		}

		if self.Config.ResponseValidator != nil && self.handlerData.shouldValidateResponse {
			log.Trace(ctx, "processing response validation handler")

			validateResponseTimer := profiling.Profile(ctx, "ValidateResponse")
			var validationFailure error
			httpStatus, validationFailure = self.Config.ResponseValidator.ValidateResponse(ctx, httpStatus, responseBody)
			validateResponseTimer.Finish()
			if err != nil {
				log.Debug(ctx, "failed response validation")
				// Validation failures are not hard errors and should be passed through to the error handler.
				// The failure is passed through since it is assumed this error contains information to be returned in the response.
				return self.processErrorResponse(ctx, responseMimeType, httpStatus, validationFailure)
			}
		}
	}

	log.Debug(ctx, "success response: %d %s", httpStatus, responseBody)

	return httpStatus, responseBody
}

func (self *Endpoint) processErrorResponse(ctx *Context, responseMimeType *MimeTypeHandler, httpStatus int, err error) (int, []byte) {
	var responseBody []byte
	var errStruct interface{}

	defer func() {
		// Print the actual error response returned to the caller.
		log.Debug(ctx, "error response: %d %s", httpStatus, responseBody)
	}()

	if httpStatus >= 500 {
		if httpStatus == 500 {
			log.Error(ctx, "failure (500): %+v", err)
		} else {
			log.Error(ctx, "failure (%d): %#v", httpStatus, err)
		}
	} else {
		log.Debug(ctx, "error (%d): %#v", httpStatus, err)
	}

	if responseMimeType == nil {
		ctx.requester.SetResponseContentType(mimeTypeTextPlain)
		responseBody = []byte(fmt.Sprintf("%#v", err))
		return httpStatus, responseBody
	}

	httpStatus, errStruct = self.Config.ErrorHandler.HandleError(ctx, httpStatus, err)
	responseBody, err = responseMimeType.Marshal(errStruct)
	if err != nil {
		ctx.requester.SetResponseContentType(mimeTypeTextPlain)
		err = errors.New(InternalCodeErrorParseFailure, "Internal server error")
		httpStatus = http.StatusInternalServerError
		responseBody = []byte(fmt.Sprintf("%s", err))
		return httpStatus, responseBody
	}

	return httpStatus, responseBody
}

func (self *Endpoint) setHandlerRequestBody(ctx *Context, mimeHandler *MimeTypeHandler, requestBody interface{}, requestBodyBytes []byte) error {
	if requestBody != nil {
		log.Trace(ctx, "non-empty request body")

		if err := mimeHandler.Unmarshal(requestBodyBytes, requestBody); err != nil {
			log.Error(ctx, "failed to unmarshal request body: %v", err)
			return ErrorRequestBodyUnmarshalFailure
		}
	}
	return nil
}

func (self *Endpoint) getHandlerResponseBody(ctx *Context, mimeHandler *MimeTypeHandler, responseBody interface{}) ([]byte, error) {
	if responseBody != nil {
		log.Trace(ctx, "non-empty response body")

		ctx.requester.SetResponseContentType(mimeHandler.MimeType)
		responseBodyBytes, err := mimeHandler.Marshal(responseBody)
		if err != nil {
			log.Error(ctx, "failed to marshal response: %v", err)
			return nil, ErrorResponseBodyMarshalFailure
		}
		if len(responseBodyBytes) == 4 && bytes.Equal(responseBodyBytes, nullBytes) {
			log.Debug(ctx, "request body is null")
			return nil, ErrorResponseBodyNull
		}
		return responseBodyBytes, nil
	}

	// No response body.
	return nil, nil
}
