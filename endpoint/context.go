package endpoint

import (
	"context"
	"time"

	"github.com/wspowell/errors"
	"github.com/wspowell/local"
	"github.com/wspowell/logging"
)

type Requester interface {
	RequestId() string

	// HTTP Method.
	Method() []byte
	// Path of the actual request URL.
	Path() []byte

	ContentType() []byte
	Accept() []byte

	VisitHeaders(f func(key []byte, value []byte))

	// MatchedPath returns the endpoint path that the request URL matches.
	// Ex: /some/path/{id}
	MatchedPath() string

	// PathParam returns the path parameter value for the given parameter name.
	// Returns false if parameter not found.
	PathParam(param string) (string, bool)
	QueryParam(param string) []byte

	RequestBody() []byte

	SetResponseContentType(contentType string)
}

var _ local.Context = (*Context)(nil)

// Context defines local endpoint data.
type Context struct {
	*local.Localized

	cancel    context.CancelFunc
	requester Requester

	HttpMethod  []byte
	MatchedPath string
}

// NewContext creates a new endpoint context. The server creates this and passes it to the endpoint handler.
// TODO: It would be really nice if *fasthttp.RequestCtx could be replaced with an interface. Not sure if this is possible.
func NewContext(serverContext context.Context, requester Requester, timeout time.Duration) *Context {
	ctx := local.FromContext(serverContext)
	cancel := local.WithTimeout(ctx, timeout)

	return &Context{
		Localized: ctx,
		cancel:    cancel,
		requester: requester,
	}
}

// Cancels the endpoint execution.
// Only call this when you need to cancel execution of child goroutines.
func (self *Context) Cancel() {
	self.cancel()
}

// ShouldContinue returns true if the underlying request has not been cancelled nor deadline exceeded.
func (self *Context) ShouldContinue() bool {
	err := self.Context.Err()

	return !(errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded))
}

// NewTestContext is an endpoint context setup for testing.
func NewTestContext(requester Requester) *Context {
	ctx := context.Background()
	serverContext := local.FromContext(ctx)

	logConfig := logging.NewConfig(logging.LevelInfo)
	logConfig.Tags()["test"] = true

	logging.WithContext(serverContext, logConfig)

	return NewContext(serverContext, requester, 30*time.Second)
}
