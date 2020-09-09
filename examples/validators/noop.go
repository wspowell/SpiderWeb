package validators

import (
	"net/http"

	"spiderweb/endpoint"
)

type NoopRequest struct {
}

func (self NoopRequest) ValidateRequest(ctx *endpoint.Context, requestBodyBytes []byte) (int, error) {
	return http.StatusOK, nil
}

type NoopResponse struct {
}

func (self NoopResponse) ValidateResponse(ctx *endpoint.Context, httpStatus int, responseBodyBytes []byte) (int, error) {
	return http.StatusOK, nil
}
