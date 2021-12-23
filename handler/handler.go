package handler

import (
	"reflect"
	"time"

	"github.com/wspowell/context"
	"github.com/wspowell/errors"
	"github.com/wspowell/log"
	"github.com/wspowell/spiderweb/body"
	"github.com/wspowell/spiderweb/httpstatus"
	"github.com/wspowell/spiderweb/mime"
	"github.com/wspowell/spiderweb/request"
	"github.com/wspowell/spiderweb/response"
)

var (
	ErrTimeout             = errors.New("request timed out")
	ErrInternalServerError = errors.New("internal server error")
)

const (
	null = "null"
)

func nullBytes() []byte {
	return []byte(null) // TODO: Is this needed?
}

type HandlerValue[T any] interface {
	*T // Essentially forces any Handler to be a value rather than a reference.
	Handler
}

type Handler interface {
	Handle(ctx context.Context) (int, error)
}

type Handle struct {
	logConfig     log.LoggerConfig
	timeout       time.Duration
	action        string
	newHandler    func() Handler
	mimeTypes     map[string]mime.Handler
	errorResponse ErrorResponse
	maxAgeSeconds int
}

func NewHandle[T any, H HandlerValue[T]](handler T) *Handle {
	return &Handle{
		action: reflect.ValueOf(handler).Type().Name(),
		newHandler: func() Handler {
			// Allocate a new handler object.
			handlerCopy := handler // Copies the Handler. Since Handler is contrained to be a *T, this SHOULD be a value, not a reference.
			return H(&handlerCopy) // Cast to a Handler.
		},
		mimeTypes: map[string]mime.Handler{
			"application/json": &mime.Json{},
		},
		errorResponse: func(statusCode int, err error) (int, []byte) {
			return statusCode, []byte(`{"error":"` + err.Error() + `"}`)
		},
	}
}

func (self *Handle) WithLogConfig(logConfig log.LoggerConfig) *Handle {
	if self.logConfig == nil {
		self.logConfig = logConfig
	}
	return self
}
func (self *Handle) LogConfig() log.LoggerConfig {
	if self.logConfig == nil {
		self.logConfig = log.NewConfig()
	}
	return self.logConfig
}

func (self *Handle) WithTimeout(timeout time.Duration) *Handle {
	if self.timeout == 0 {
		self.timeout = timeout
	}
	return self
}

func (self *Handle) Timeout() time.Duration {
	if self.timeout == 0 {
		return 30 * time.Second
	}
	return self.timeout
}

func (self *Handle) WithETag(maxAgeSeconds int) *Handle {
	self.maxAgeSeconds = maxAgeSeconds
	return self
}

func (self *Handle) WithMimeTypes(mimeTypes map[string]mime.Handler) *Handle {
	for mimeType, mimeTypeHandler := range mimeTypes {
		if _, exists := self.mimeTypes[mimeType]; !exists {
			self.mimeTypes[mimeType] = mimeTypeHandler
		}
	}
	return self
}

type ErrorResponse func(statusCode int, err error) (int, []byte)

func (self *Handle) WithErrorResponse(errorResponse ErrorResponse) *Handle {
	self.errorResponse = errorResponse
	return self
}

func (self *Handle) Run(ctx context.Context, requester request.Requester) (int, []byte) {
	ctx, cancel := context.WithTimeout(ctx, self.Timeout())
	defer cancel()

	// Every invocation of an endpoint creates its own logger instance.
	ctx = log.WithContext(ctx, self.LogConfig())

	var statusCode int
	var responseBytes []byte

	err := errors.Catch(func() {
		statusCode, responseBytes = self.run(ctx, requester)
	})

	if err != nil {
		log.Error(ctx, "%s", err)
		log.Debug(ctx, "%+v", err)

		return self.errorResponse(httpstatus.InternalServerError, errors.Wrap(err, ErrInternalServerError))
	}

	return statusCode, responseBytes
}

func (self *Handle) run(ctx context.Context, requester request.Requester) (int, []byte) {
	var err error

	if err = ctx.Err(); err != nil {
		return self.errorResponse(httpstatus.RequestTimeout, errors.Wrap(err, ErrTimeout))
	}

	handlerInstance := self.newHandler()

	self.setLogTags(ctx, requester)
	requester.SetResponseContentType(string(requester.Accept())) // FIXME: should set a default for when not set on request.
	requester.SetResponseHeader("X-Request-Id", requester.RequestId())

	if asBodyRequester, ok := any(handlerInstance).(body.Requester); ok {
		if err = self.processRequest(requester, asBodyRequester); err != nil {
			return self.errorResponse(httpstatus.InternalServerError, err)
		}
	}

	if err = self.processParameters(ctx, requester, handlerInstance); err != nil {
		return self.errorResponse(httpstatus.InternalServerError, err)
	}

	if err = ctx.Err(); err != nil {
		return self.errorResponse(httpstatus.RequestTimeout, errors.Wrap(err, ErrTimeout))
	}

	// Run the endpoint
	var statusCode int
	if statusCode, err = handlerInstance.Handle(ctx); err != nil {
		return self.errorResponse(statusCode, err)
	}

	if err = ctx.Err(); err != nil {
		return self.errorResponse(httpstatus.RequestTimeout, errors.Wrap(err, ErrTimeout))
	}

	// Handle the response.
	var responseBytes []byte
	if asBodyResponder, ok := any(handlerInstance).(body.Responder); ok {
		if responseBytes, err = self.processResponse(requester, asBodyResponder); err != nil {
			return self.errorResponse(httpstatus.InternalServerError, err)
		}
	}

	if self.maxAgeSeconds != 0 {
		log.Trace(ctx, "eTagEnabled, handling etag")
		return response.HandleETag(ctx, requester, self.maxAgeSeconds, statusCode, responseBytes)
	}

	return statusCode, responseBytes
}

func (self *Handle) setLogTags(ctx context.Context, requester request.Requester) {
	log.Tag(ctx, "request_id", requester.RequestId())
	log.Tag(ctx, "method", string(requester.Method()))
	log.Tag(ctx, "route", requester.MatchedPath())
	log.Tag(ctx, "path", string(requester.Path()))
	log.Tag(ctx, "action", self.action)
}

func (self *Handle) processRequest(requester request.Requester, e body.Requester) error {
	if mimeTypeHandler, exists := self.mimeTypes[string(requester.ContentType())]; exists {
		return e.UnmarshalRequestBody(requester.RequestBody(), mimeTypeHandler)
	}
	return errors.New("Content-Type mime type not supported: %s", requester.ContentType) // FIXME: wrap error
}

func (self *Handle) processResponse(requester request.Requester, e body.Responder) ([]byte, error) {
	if mimeTypeHandler, exists := self.mimeTypes[string(requester.Accept())]; exists {
		return e.MarshalResponseBody(mimeTypeHandler)
	}
	return nil, errors.New("Accept mime type not supported: %s", requester.ContentType) // FIXME: wrap error
}

func (self *Handle) processParameters(ctx context.Context, requester request.Requester, e any) error {
	if asPathParams, ok := e.(request.PathParameters); ok {
		for _, pathParam := range asPathParams.PathParameters() {
			if pathParamValue, exists := requester.PathParam(pathParam.ParamName()); exists {
				if err := pathParam.SetParam(pathParamValue); err != nil {
					return err // FIXME: wrap error
				}
				// Each path parameter is added as a log tag.
				// Note: It helps if the path parameter name is descriptive.
				log.Tag(ctx, pathParam.ParamName(), pathParamValue)
				continue
			}
			return errors.New("request does not have path parameter: %s", pathParam.ParamName()) // FIXME: wrap error
		}
	}

	if asQueryParams, ok := e.(request.QueryParameters); ok {
		for _, queryParam := range asQueryParams.QueryParameters() {
			if queryParamValue, exists := requester.QueryParam(queryParam.ParamName()); exists {
				if err := queryParam.SetParam(string(queryParamValue)); err != nil {
					return err // FIXME: wrap error
				}
				continue
			}
			return errors.New("request does not have path parameter: %s", queryParam.ParamName()) // FIXME: wrap error
		}
	}

	return nil
}
