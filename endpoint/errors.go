package endpoint

import "github.com/wspowell/errors"

var (
	ErrInternalServerError = errors.New("internal server error")
	ErrBadRequest          = errors.New("bad request")
	ErrInvalidBody         = errors.New("invalid body")
	ErrRequestTimeout      = errors.New("request timeout")
	ErrInvalidMimeType     = errors.New("invalid MIME type")
)
