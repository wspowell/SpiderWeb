package handler

import (
	"github.com/wspowell/context"
	"github.com/wspowell/errors"
	"github.com/wspowell/spiderweb/body"
	"github.com/wspowell/spiderweb/endpoint"
	"github.com/wspowell/spiderweb/httpstatus"
	"github.com/wspowell/spiderweb/mime"
	"github.com/wspowell/spiderweb/request"
)

type HandlerValue[T any] interface {
	*T // Essentially forces any Handler to be a value rather than a reference.
	Handler
}

type Handler interface {
	body.Responder
	Handle(ctx context.Context) (int, error)
}

type Handle struct {
	newHandler    func() Handler
	mimeTypes     map[string]mime.Handler
	errorResponse ErrorResponse
}

func NewHandle[T any, H HandlerValue[T]](handler T) *Handle {
	return &Handle{
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

func (self *Handle) Run(ctx context.Context, request endpoint.Requester) (int, []byte) {
	var err error
	handlerInstance := self.newHandler()

	if asBodyRequester, ok := any(handlerInstance).(body.Requester); ok {
		if err = self.processRequest(request, asBodyRequester); err != nil {
			return self.errorResponse(httpstatus.InternalServerError, err)
		}
	}

	if err = self.processParameters(request, handlerInstance); err != nil {
		return self.errorResponse(httpstatus.InternalServerError, err)
	}

	// Run the endpoint
	var statusCode int
	if statusCode, err = handlerInstance.Handle(ctx); err != nil {
		return self.errorResponse(statusCode, err)
	}

	// Handle the response.
	var responseBytes []byte
	if responseBytes, err = self.processResponse(request, handlerInstance); err != nil {
		return self.errorResponse(httpstatus.InternalServerError, err)
	}

	return statusCode, responseBytes
}

func (self *Handle) processRequest(request endpoint.Requester, e body.Requester) error {
	if mimeTypeHandler, exists := self.mimeTypes[string(request.ContentType())]; exists {
		return e.UnmarshalRequestBody(request.RequestBody(), mimeTypeHandler)
	}
	return errors.New("Content-Type mime type not supported: %s", request.ContentType) // FIXME: wrap error
}

func (self *Handle) processResponse(request endpoint.Requester, e body.Responder) ([]byte, error) {
	if mimeTypeHandler, exists := self.mimeTypes[string(request.Accept())]; exists {
		return e.MarshalResponseBody(mimeTypeHandler)
	}
	return nil, errors.New("Accept mime type not supported: %s", request.ContentType) // FIXME: wrap error
}

func (self *Handle) processParameters(req endpoint.Requester, e any) error {
	if asPathParams, ok := e.(request.PathParameters); ok {
		for _, pathParam := range asPathParams.PathParameters() {
			if pathParamValue, exists := req.PathParam(pathParam.ParamName()); exists {
				if err := pathParam.SetParam(pathParamValue); err != nil {
					return err // FIXME: wrap error
				}
				continue
			}
			return errors.New("request does not have path parameter: %s", pathParam.ParamName()) // FIXME: wrap error
		}
	}

	if asQueryParams, ok := e.(request.QueryParameters); ok {
		for _, queryParam := range asQueryParams.QueryParameters() {
			if queryParamValue, exists := req.QueryParam(queryParam.ParamName()); exists {
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
