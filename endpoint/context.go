package endpoint

import (
	"github.com/wspowell/context"
	"github.com/wspowell/errors"
)

// Context defines local endpoint data.
type Context struct {
	context.Context

	requester Requester
}

// NewContext creates a new endpoint context. The server creates this and passes it to the endpoint handler.
func NewContext(serverContext context.Context, requester Requester) *Context {
	ctx := context.Localize(serverContext)

	return &Context{
		Context:   ctx,
		requester: requester,
	}
}

// ShouldContinue returns true if the underlying request has not been cancelled nor deadline exceeded.
func (self *Context) ShouldContinue() bool {
	err := self.Context.Err()

	return !(errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded))
}
