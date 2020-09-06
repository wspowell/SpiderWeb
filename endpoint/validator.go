package endpoint

import (
	"net/http"
)

type RequestValidator interface {
	ValidateRequest(ctx *Context, requestBodyBytes []byte) (int, error)
}

type RequestValidation struct{}

func (self RequestValidation) ValidateRequest(ctx *Context, requestBodyBytes []byte) (int, error) {
	return http.StatusOK, nil
}

type ResponseValidator interface {
	ValidateResponse(ctx *Context, httpStatus int, responseBodyBytes []byte) (int, error)
}

type ResponseValidation struct{}

func (self ResponseValidation) ValidateResponse(ctx *Context, httpStatus int, responseBodyBytes []byte) (int, error) {
	return http.StatusOK, nil
}
