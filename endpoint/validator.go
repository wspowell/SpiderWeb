package endpoint

type RequestValidator interface {
	// ValidateRequest and return validation failures.
	// Errors returned are passed straight through to the ErrorHandler.
	// Returned status code is not used unless error is not nil.
	ValidateRequest(ctx *Context, requestBody []byte) (int, error)
}

type ResponseValidator interface {
	// ValidateRequest and return validation failures.
	// Errors returned are passed straight through to the ErrorHandler.
	// Returned status code is not used unless error is not nil.
	ValidateResponse(ctx *Context, httpStatus int, responseBody []byte) (int, error)
}
