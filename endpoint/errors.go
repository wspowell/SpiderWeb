package endpoint

const (
	internalServerError = "internal server error"
	badRequest          = "bad request"
)

// Internal Codes.

// Internal error codes.
const (
	icPanic             = "SW000"
	icErrorParseFailure = "SW001"
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
	icRequestTimeout3             = "SW192"
	icRequestTimeout4             = "SW193"
	icRequestTimeout5             = "SW194"
)

// Response errors codes.
const (
	icResponseMimeTypeUnsupported = "SW201"
	icResponseMimeTypeMissing     = "SW202"
	icResponseBodyMarshalFailure  = "SW203"
	icResponseBodyNull            = "SW204"
)

// Reflection error codes.
const (
	icPathParamCannotSet       = "SW901"
	icPathParamValueNotFound   = "SW902"
	icCannotSetValueFromString = "SW903"
	icPathParamSetFailure      = "SW904"
	icQueryParamCannotSet      = "SW905"
	icQueryParamValueNotFound  = "SW906"
	icQueryParamSetFailure     = "SW907"
	icResourceNotSet           = "SW908"
	icResourceNotValid         = "SW909"
)
