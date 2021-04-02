package endpoint

import (
	"github.com/wspowell/errors"
)

// All errors emitted by endpoint.
// It is expected that consumers check for FrameworkErrors in their ErrorHandlers and process them accordingly.
var (
	ErrorPanic = errors.New("SW0", "internal server error")

	ErrorRequestTimeout              = errors.New("SW1", "request timeout")
	ErrorRequestUnsupportedMimeType  = errors.New("SW2", "unsupported request MIME type")
	ErrorRequestBodyReadFailure      = errors.New("SW3", "request body read failure")
	ErrorRequestBodyUnmarshalFailure = errors.New("SW4", "request body unmarshal failure")
	ErrorRequestValidationError      = errors.New("SW5", "request validation error")

	ErrorResponseBodyMarshalFailure  = errors.New("SW6", "response body marshal failure")
	ErrorResponseBodyMissing         = errors.New("SW7", "missing response body")
	ErrorResponseBodyNull            = errors.New("SW8", "response body null")
	ErrorResponseUnsupportedMimeType = errors.New("SW9", "unsupported response MIME type")
)

const (
	InternalCodeRequestMimeTypeMissing      = "SW1"
	InternalCodeRequestMimeTypeUnsupported  = "SW2"
	InternalCodeResponseMimeTypeUnsupported = "SW3"
	InternalCodeResponseMimeTypeMissing     = "SW4"
)
