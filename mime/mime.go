package mime

import (
	"github.com/wspowell/errors"
)

var (
	ErrNotSupported = errors.New("mime type not supported")
	ErrMarshal      = errors.New("mime type could not be marshaled")
	ErrUnmarshal    = errors.New("mime type could not be unmarshaled")
)

type Handler interface {
	Unmarshaler
	Marshaler
}

type Unmarshaler interface {
	// UnmarshalMimeType `data` into `value` for parsing a request body.
	// This function MUST check if the `value` implements the mime type behavior.
	// If `value` does not implement the mime type, then the function MUST return ErrNotSupported.
	// If `value` does implement the mime type, then it MUST unmarshal `data` into `value`.
	// Errors:
	//   * ErrUnmarshal
	//   * ErrNotSupported
	UnmarshalMimeType(data []byte, value any) error
}

type Marshaler interface {
	// MarshalMimeType into bytes for responding to a request.
	// This function MUST check if the `value` implements the mime type behavior.
	// If `value` does not implement the mime type, then the function MUST return ErrNotSupported.
	// If `value` does implement the mime type, then it MUST marshal `value` and return the `[]byte`.
	// Errors:
	//   * ErrMarshal
	//   * ErrNotSupported
	MarshalMimeType(value any) ([]byte, error)
}
