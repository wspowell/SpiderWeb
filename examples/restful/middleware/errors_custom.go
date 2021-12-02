package middleware

import (
	"github.com/wspowell/context"
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

func (self ErrorJsonWithCodeResponse) HandleError(ctx context.Context, httpStatus int, err error) (int, interface{}) {
	var myErr ErrorWithCodes
	if errors.As(err, &myErr) {
		return httpStatus, ErrorJsonWithCodeResponse{
			Code:    myErr.Code,
			Message: myErr.Message,
		}
	}

	// Anything not using ErrorWithCodes.
	return httpStatus, ErrorJsonWithCodeResponse{
		Code:    "INTERNAL_ERROR",
		Message: err.Error(),
	}
}
