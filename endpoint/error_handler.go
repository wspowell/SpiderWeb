package endpoint

import (
	"fmt"

	"github.com/wspowell/context"
)

type ErrorHandler interface {

	// FIXME: HandlerError only really converts an error into []byte. This function definition could be made simpler.
	//        The error could be returned and then the internal endpoint code handles the marshaling.
	HandleError(ctx context.Context, httpStatus int, err error) (int, interface{})
}

type defaultErrorResponse struct {
	Message string `json:"message"`
}

type defaultErrorHandler struct{}

func (self defaultErrorHandler) HandleError(ctx context.Context, httpStatus int, err error) (int, interface{}) {
	return httpStatus, defaultErrorResponse{
		Message: fmt.Sprintf("%v", err),
	}
}
