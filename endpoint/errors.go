package endpoint

import (
	"fmt"

	"spiderweb/errors"
)

// All errors emitted by endpoint.
// It is expected that consumers check for FrameworkErrors in their ErrorHandlers and process them accordingly.
var (
	// FrameworkError is the base error and all other framework errors wrap this one.
	FrameworkError = fmt.Errorf("spiderweb framework error")

	ErrorPanic = wrapFrameworkError("internal server error")

	ErrorRequestTimeout              = wrapFrameworkError("request timeout")
	ErrorRequestUnknownMimeType      = wrapFrameworkError("unknown request MIME type")
	ErrorRequestBodyReadFailure      = wrapFrameworkError("request body read failure")
	ErrorRequestBodyUnmarshalFailure = wrapFrameworkError("request body unmarshal failure")
	ErrorRequestValidationError      = wrapFrameworkError("request validation error")

	ErrorResponseBodyMarshalFailure = wrapFrameworkError("response body marshal failure")
	ErrorResponseBodyMissing        = wrapFrameworkError("missing response body")
	ErrorResponseBodyNull           = wrapFrameworkError("response body null")
	ErrorResponseUnknownMimeType    = wrapFrameworkError("unknown response MIME type")
)

func wrapFrameworkError(message string) error {
	return fmt.Errorf("%w: %s", FrameworkError, message)
}

// HasFrameworkError returns true if any error is returned from the spiderweb framework.
func HasFrameworkError(err error) bool {
	return errors.Is(err, FrameworkError)
}
