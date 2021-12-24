package handler

import (
	"github.com/wspowell/errors"
	"github.com/wspowell/spiderweb/request"
)

var (
	ErrInvalidAuthFormat = errors.New("invalid auth format")
	ErrUnauthorized      = errors.New("user not authorized")
)

var (
	AuthBasic = []byte("Basic ")
)

type Authorizer interface {
	Authorize(requester request.Requester) error
}
