package endpoint

import (
	"fmt"

	"github.com/wspowell/errors"
)

// All errors emitted by endpoint.
// It is expected that consumers check for FrameworkErrors in their ErrorHandlers and process them accordingly.
var (
	// FrameworkError is the base error and all other framework errors wrap this one.
	FrameworkError = fmt.Errorf("spiderweb framework error")

	ErrorPanic = wrapFrameworkError("internal server error")

	ErrorRequestTimeout              = wrapFrameworkError("request timeout")
	ErrorRequestUnsupportedMimeType  = wrapFrameworkError("unsupported request MIME type")
	ErrorRequestBodyReadFailure      = wrapFrameworkError("request body read failure")
	ErrorRequestBodyUnmarshalFailure = wrapFrameworkError("request body unmarshal failure")
	ErrorRequestValidationError      = wrapFrameworkError("request validation error")

	ErrorResponseBodyMarshalFailure  = wrapFrameworkError("response body marshal failure")
	ErrorResponseBodyMissing         = wrapFrameworkError("missing response body")
	ErrorResponseBodyNull            = wrapFrameworkError("response body null")
	ErrorResponseUnsupportedMimeType = wrapFrameworkError("unsupported response MIME type")
)

const (
	InternalCodeRequestMimeTypeMissing      = "SW1"
	InternalCodeRequestMimeTypeUnsupported  = "SW2"
	InternalCodeResponseMimeTypeUnsupported = "SW3"
	InternalCodeResponseMimeTypeMissing     = "SW4"
)

func wrapFrameworkError(message string) error {
	return fmt.Errorf("%w: %s", FrameworkError, message)
}

// HasFrameworkError returns true if any error is returned from the spiderweb framework.
func HasFrameworkError(err error) bool {
	return errors.Is(err, FrameworkError)
}
