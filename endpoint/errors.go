package endpoint

import "github.com/wspowell/errors"

var (
	ErrInternalServerError = errors.New(icInternalServerError, "internal server error")
	ErrBadRequest          = errors.New(icBadRequest, "bad request")
)

// Internal Codes.

// Internal error codes.
const (
	icPanic                    = "SW000"
	icInternalServerError      = "SW001"
	icBadRequest               = "SW002"
	icErrorParseFailure        = "SW002"
	icAuthorizerInterfaceError = "SW003"
)

// Request errors codes.
const (
	icRequestMimeTypeMissing      = "SW100"
	icRequestMimeTypeUnsupported  = "SW101"
	icRequestBodyUnmarshalFailure = "SW102"
	icRequestPathParamsError      = "SW103"
	icRequestQueryParamsError     = "SW104"
	icRequestResourcesError       = "SW105"
	icRequestTimeout1             = "SW190"
	icRequestTimeout2             = "SW191"
)

// Response errors codes.
const (
	icResponseMimeTypeUnsupported = "SW201"
	icResponseMimeTypeMissing     = "SW202"
	icResponseBodyMarshalFailure  = "SW203"
	icResponseBodyNull            = "SW204"
)

const (
	icNewHttpRequesterReadAllError = "SW300"
)

// Reflection error codes.
const (
	icPathParamCannotSet             = "SW901"
	icPathParamValueNotFound         = "SW902"
	icCannotSetValueFromString       = "SW903"
	icPathParamSetFailure            = "SW904"
	icQueryParamCannotSet            = "SW905"
	icQueryParamValueNotFound        = "SW906"
	icQueryParamSetFailure           = "SW907"
	icResourceNotSet                 = "SW908"
	icResourceNotValid               = "SW909"
	icInvalidTypeForStringConversion = "SW910"
)
