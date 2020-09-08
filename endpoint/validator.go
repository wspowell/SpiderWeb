package endpoint

import (
	"net/http"
)

type RequestValidator interface {
	// ValidateRequest and return validation failures.
	// Errors returned are passed straight through to the ErrorHandler.
	ValidateRequest(ctx *Context, requestBodyBytes []byte) (int, error)
}

type RequestValidation struct{}

func (self RequestValidation) ValidateRequest(ctx *Context, requestBodyBytes []byte) (int, error) {
	return http.StatusOK, nil
}

type ResponseValidator interface {
	// ValidateRequest and return validation failures.
	// Errors returned are passed straight through to the ErrorHandler.
	ValidateResponse(ctx *Context, httpStatus int, responseBodyBytes []byte) (int, error)
}

type ResponseValidation struct{}

func (self ResponseValidation) ValidateResponse(ctx *Context, httpStatus int, responseBodyBytes []byte) (int, error) {
	return http.StatusOK, nil
}
