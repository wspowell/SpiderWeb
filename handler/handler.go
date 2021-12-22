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

type Handler[T any] interface {
	*T // Essentially forces any Handler to be a value rather than a reference.
	body.Responder
	Handle(ctx context.Context) (int, error)
}

type Runner func(ctx context.Context, request endpoint.Requester, mimeTypes map[string]mime.Handler) (int, []byte)

func New[T any, H Handler[T]](handler T) Runner {
	return func(ctx context.Context, request endpoint.Requester, mimeTypes map[string]mime.Handler) (int, []byte) {
		var err error

		// Allocate a new handler object.
		handlerCopy := handler             // Copies the Handler. Since it is contrained to be a *T, this SHOULD be a value, not a reference.
		handlerInstance := H(&handlerCopy) // Cast to a Handler.

		if asBodyRequester, ok := any(handlerInstance).(body.Requester); ok {
			if err = ProcessRequest(request, asBodyRequester, mimeTypes); err != nil {
				return httpstatus.InternalServerError, errorToBytes(err)
			}
		}

		if err = ProcessParameters(request, handlerInstance); err != nil {
			return httpstatus.InternalServerError, errorToBytes(err)
		}

		// Run the endpoint
		var statusCode int
		statusCode, err = handlerInstance.Handle(ctx)

		// Handle the response.
		var responseBytes []byte
		if responseBytes, err = ProcessResponse(request, handlerInstance, mimeTypes); err != nil {
			return httpstatus.InternalServerError, errorToBytes(err)
		}

		return statusCode, responseBytes
	}
}

// FIXME: this needs to be a better handler that formats into a valid body
func errorToBytes(err error) []byte {
	return []byte(err.Error())
}

type Request struct {
	Body        []byte
	ContentType string
	Accept      string
	PathParams  map[string]string
	QueryParams map[string]string
}

func ProcessRequest(request endpoint.Requester, e body.Requester, mimeTypes map[string]mime.Handler) error {
	if mimeTypeHandler, exists := mimeTypes[string(request.ContentType())]; exists {
		return e.UnmarshalRequestBody(request.RequestBody(), mimeTypeHandler)
	}
	return errors.New("Content-Type mime type not supported: %s", request.ContentType) // FIXME: wrap error
}

func ProcessResponse(request endpoint.Requester, e body.Responder, mimeTypes map[string]mime.Handler) ([]byte, error) {
	if mimeTypeHandler, exists := mimeTypes[string(request.Accept())]; exists {
		return e.MarshalResponseBody(mimeTypeHandler)
	}
	return nil, errors.New("Accept mime type not supported: %s", request.ContentType) // FIXME: wrap error
}

func ProcessParameters(req endpoint.Requester, e any) error {
	if asPathParams, ok := e.(interface{ PathParameters() []request.Parameter }); ok {
		for _, pathParam := range asPathParams.PathParameters() {
			if pathParamValue, exists := req.PathParam(pathParam.Name()); exists {
				if err := pathParam.SetParam(pathParamValue); err != nil {
					return err // FIXME: wrap error
				}
				continue
			}
			return errors.New("request does not have path parameter: %s", pathParam.Name()) // FIXME: wrap error
		}
	}

	if asQueryParams, ok := e.(interface{ QueryParameters() []request.Parameter }); ok {
		for _, queryParam := range asQueryParams.QueryParameters() {
			if queryParamValue, exists := req.QueryParam(queryParam.Name()); exists {
				if err := queryParam.SetParam(string(queryParamValue)); err != nil {
					return err // FIXME: wrap error
				}
				continue
			}
			return errors.New("request does not have path parameter: %s", queryParam.Name()) // FIXME: wrap error
		}
	}

	return nil
}
