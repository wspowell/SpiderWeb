package handler

import (
	"reflect"
	"time"

	"github.com/wspowell/context"
	"github.com/wspowell/errors"
	"github.com/wspowell/log"
	"github.com/wspowell/spiderweb/mime"
)

var (
	ErrTimeout             = errors.New("request timed out")
	ErrInternalServerError = errors.New("internal server error")
)

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
	maxAgeSeconds time.Duration
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
		errorResponse: func(statusCode *int, responseBody *[]byte, err error) {
			*responseBody = (*responseBody)[:0]
			*responseBody = append(*responseBody, []byte(`{"error":"`+err.Error()+`"}`)...)
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

func (self *Handle) WithETag(maxAgeSeconds time.Duration) *Handle {
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

func (self *Handle) MimeTypes() map[string]mime.Handler {
	mimeTypes := make(map[string]mime.Handler, len(self.mimeTypes))
	for mimeType, mimeTypeHandler := range self.mimeTypes {
		mimeTypes[mimeType] = mimeTypeHandler
	}
	return mimeTypes
}

type ErrorResponse func(statusCode *int, responseBody *[]byte, err error)

func (self *Handle) WithErrorResponse(errorResponse ErrorResponse) *Handle {
	self.errorResponse = errorResponse
	return self
}

func (self *Handle) Runner() *Runner {
	return &Runner{
		Handle: Handle{
			logConfig:     self.LogConfig(),
			timeout:       self.Timeout(),
			action:        self.action,
			newHandler:    self.newHandler,
			mimeTypes:     self.MimeTypes(),
			errorResponse: self.errorResponse,
			maxAgeSeconds: self.maxAgeSeconds,
		},
	}
}
