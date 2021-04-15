package endpoint

import (
	"context"
	"time"

	"github.com/wspowell/errors"
	"github.com/wspowell/local"
	"github.com/wspowell/logging"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

var _ local.Context = (*Context)(nil)

// Context defines local endpoint data.
type Context struct {
	*local.Localized

	cancel     context.CancelFunc
	requestCtx *fasthttp.RequestCtx

	HttpMethod  []byte
	MatchedPath string
}

// NewContext creates a new endpoint context. The server creates this and passes it to the endpoint handler.
// TODO: It would be really nice if *fasthttp.RequestCtx could be replaced with an interface. Not sure if this is possible.
func NewContext(serverContext context.Context, requestCtx *fasthttp.RequestCtx, timeout time.Duration) *Context {
	ctx := local.FromContext(serverContext)
	cancel := local.WithTimeout(ctx, timeout)

	var matchedPath string
	matchedPath, _ = requestCtx.UserValue(router.MatchedRoutePathParam).(string)

	return &Context{
		Localized:   ctx,
		cancel:      cancel,
		requestCtx:  requestCtx,
		HttpMethod:  requestCtx.Method(),
		MatchedPath: matchedPath,
	}
}

// Request returns the current request.
func (self *Context) Request() *fasthttp.Request {
	return &self.requestCtx.Request
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
func NewTestContext() *Context {
	ctx := context.Background()
	serverContext := local.FromContext(ctx)

	logConfig := logging.NewConfig(logging.LevelInfo)
	logConfig.Tags()["test"] = true

	logging.WithContext(serverContext, logConfig)

	requestCtx := fasthttp.RequestCtx{}
	requestCtx.Init(&fasthttp.Request{}, nil, nil)

	return NewContext(serverContext, &requestCtx, 30*time.Second)
}
