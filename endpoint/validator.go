package endpoint

type RequestValidator interface {
	// ValidateRequest and return validation failures.
	// Errors returned are passed straight through to the ErrorHandler.
	ValidateRequest(ctx *Context, requestBody []byte) (int, error)
}

type ResponseValidator interface {
	// ValidateRequest and return validation failures.
	// Errors returned are passed straight through to the ErrorHandler.
	ValidateResponse(ctx *Context, httpStatus int, responseBody []byte) (int, error)
}
