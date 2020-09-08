package error_handlers

import (
	"encoding/json"
	"net/http"

	"spiderweb/endpoint"
	"spiderweb/errors"
)

type ErrorWithCodes struct {
	Code         string
	InternalCode string
	Message      string
}

func NewErrorWithCodes(code string, message string) error {
	return ErrorWithCodes{
		Code:    code,
		Message: message,
	}
}

func (self ErrorWithCodes) Error() string {
	return self.Message
}

var _ endpoint.ErrorHandler = (*ErrorJsonWithCodeResponse)(nil)

type ErrorJsonWithCodeResponse struct {
	Code         string `json:"code"`
	InternalCode string `json:"internal_code"`
	Message      string `json:"message"`
}

func (self ErrorJsonWithCodeResponse) HandleError(ctx *endpoint.Context, httpStatus int, err error) (int, []byte) {
	var errorBytes []byte
	var responseErr error

	var myErr ErrorWithCodes
	if errors.As(err, &myErr) {
		errorBytes, responseErr = json.Marshal(ErrorJsonWithCodeResponse{
			Code:         myErr.Code,
			InternalCode: errors.InternalCode(err),
			Message:      myErr.Message,
		})
	} else {
		// Catch anything not using ErrorWithCodes.
		errorBytes, responseErr = json.Marshal(ErrorJsonWithCodeResponse{
			Code:         "INTERNAL_ERROR",
			InternalCode: errors.InternalCode(err),
			Message:      err.Error(),
		})
	}

	if responseErr != nil {
		// Provide a valid default for responding.
		httpStatus = http.StatusInternalServerError
		errorBytes = []byte(`{"code":"INTERNAL_ERROR","internal_code":"SW0000","message":"internal server error"}`)
	}

	return httpStatus, errorBytes
}
