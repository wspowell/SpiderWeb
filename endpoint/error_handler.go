package endpoint

import (
	"fmt"

	"github.com/wspowell/context"
)

type ErrorHandler interface {
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
