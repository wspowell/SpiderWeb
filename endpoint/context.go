package endpoint

import (
	"spiderweb/local"
	"spiderweb/logging"

	"github.com/valyala/fasthttp"
)

var _ local.Context = (*Context)(nil)

type Context struct {
	*local.Localized
	logging.Loggerer

	requestCtx *fasthttp.RequestCtx
}

// TODO: It would be really nice if *fasthttp.RequestCtx could be replaced with an interface.
func NewContext(requestCtx *fasthttp.RequestCtx, logger logging.Loggerer) *Context {
	return &Context{
		// FIXME: This makes the request context into the context. This is probably not thread safe.
		Localized:  local.FromContext(requestCtx),
		Loggerer:   logger,
		requestCtx: requestCtx,
	}
}

func (self *Context) Request() *fasthttp.Request {
	return &self.requestCtx.Request
}
