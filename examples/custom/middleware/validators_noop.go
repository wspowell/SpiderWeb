package middleware

import (
	"net/http"

	"github.com/wspowell/spiderweb/endpoint"
)

type ValidateNoopRequest struct {
}

func (self ValidateNoopRequest) ValidateRequest(ctx *endpoint.Context, requestBodyBytes []byte) (int, error) {
	return http.StatusOK, nil
}

type ValidateNoopResponse struct {
}

func (self ValidateNoopResponse) ValidateResponse(ctx *endpoint.Context, httpStatus int, responseBodyBytes []byte) (int, error) {
	return http.StatusOK, nil
}
