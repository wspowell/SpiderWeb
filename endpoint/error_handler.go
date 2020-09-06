package endpoint

import (
	"encoding/json"
	"net/http"

	"spiderweb/errors"
)

type ErrorHandler interface {
	HandleError(ctx *Context, httpStatus int, err error) (int, []byte)
}

// ErrorResponse error handler default.
type ErrorResponse struct {
	InternalCode string `json:"internal_code"`
	Message      string `json:"message"`
}

func (self ErrorResponse) HandleError(ctx *Context, httpStatus int, err error) (int, []byte) {
	var errorBytes []byte
	var responseErr error

	errorBytes, responseErr = json.Marshal(ErrorResponse{
		InternalCode: errors.InternalCode(err),
		Message:      err.Error(),
	})

	if responseErr != nil {
		// Provide a valid default for responding.
		httpStatus = http.StatusInternalServerError
		errorBytes = []byte(`{"internal_code":"SW0000","message":"internal server error"}`)
	}

	return httpStatus, errorBytes
}
