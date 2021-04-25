package endpoint

import "github.com/wspowell/context"

type RequestValidator interface {
	// ValidateRequest and return validation failures.
	// Errors returned are passed straight through to the ErrorHandler.
	// Returned status code is not used unless error is not nil.
	ValidateRequest(ctx context.Context, requestBody []byte) (int, error)
}

type ResponseValidator interface {
	// ValidateRequest and return validation failures.
	// Errors returned are passed straight through to the ErrorHandler.
	// Returned status code is not used unless error is not nil.
	ValidateResponse(ctx context.Context, httpStatus int, responseBody []byte) (int, error)
}
