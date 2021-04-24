package endpoint

import (
	"context"

	"github.com/wspowell/errors"
	"github.com/wspowell/local"
)

var _ local.Context = (*Context)(nil)

// Context defines local endpoint data.
type Context struct {
	*local.Localized

	requester Requester
}

// NewContext creates a new endpoint context. The server creates this and passes it to the endpoint handler.
func NewContext(serverContext context.Context, requester Requester) *Context {
	ctx := local.FromContext(serverContext)

	return &Context{
		Localized: ctx,
		requester: requester,
	}
}

// ShouldContinue returns true if the underlying request has not been cancelled nor deadline exceeded.
func (self *Context) ShouldContinue() bool {
	err := self.Context.Err()

	return !(errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded))
}
