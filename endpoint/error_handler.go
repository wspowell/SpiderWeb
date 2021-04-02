package endpoint

import (
	"bytes"
	"fmt"
)

type ErrorHandler interface {
	MimeType() string
	HandleError(ctx *Context, httpStatus int, err error) (int, []byte)
}

type DefaultErrorResponse struct {
	Message string `json:"message"`
}

type defaultErrorHandler struct{}

func (self defaultErrorHandler) MimeType() string {
	return mimeTypeTextPlain
}

func (self defaultErrorHandler) HandleError(ctx *Context, httpStatus int, err error) (int, []byte) {
	return httpStatus, bytes.NewBufferString(fmt.Sprintf("%#v", err)).Bytes()
}
