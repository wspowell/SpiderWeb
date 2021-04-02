package error_handlers

import (
	"encoding/json"

	"github.com/wspowell/errors"
	"github.com/wspowell/spiderweb/endpoint"
)

type ErrorWithCodes struct {
	Code    string
	Message string
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

func (self ErrorJsonWithCodeResponse) MimeType() string {
	return "application/json"
}

func (self ErrorJsonWithCodeResponse) HandleError(ctx *endpoint.Context, httpStatus int, err error) (int, []byte) {
	var errorBytes []byte

	var myErr ErrorWithCodes
	if errors.As(err, &myErr) {
		errorBytes, _ = json.Marshal(ErrorJsonWithCodeResponse{
			Code:    myErr.Code,
			Message: myErr.Message,
		})
	} else {
		// Catch anything not using ErrorWithCodes.
		errorBytes, _ = json.Marshal(ErrorJsonWithCodeResponse{
			Code:    "INTERNAL_ERROR",
			Message: err.Error(),
		})
	}

	return httpStatus, errorBytes
}
