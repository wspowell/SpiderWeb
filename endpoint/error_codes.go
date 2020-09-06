package endpoint

// Error codes.
const (
	ErrorCodePanic                           = "SW0001"
	ErrorCodeMissingResponseBody             = "SW0002"
	ErrorCodeRequestBodyCopyFailure          = "SW0003"
	ErrorCodeValidationError                 = "SW0004"
	ErrorCodeResponseBodyJsonMarshalFailure  = "SW0005"
	ErrorCodeRequestBodyJsonUnmarshalFailure = "SW0006"
	ErrorCodeResponseBodyNull                = "SW0007"
)
