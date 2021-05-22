package middleware

import (
	"net/http"

	"github.com/wspowell/context"
)

type ValidateNoopRequest struct {
}

func (self ValidateNoopRequest) ValidateRequest(ctx context.Context, requestBodyBytes []byte) (int, error) {
	return http.StatusOK, nil
}

type ValidateNoopResponse struct {
}

func (self ValidateNoopResponse) ValidateResponse(ctx context.Context, httpStatus int, responseBodyBytes []byte) (int, error) {
	return http.StatusOK, nil
}
