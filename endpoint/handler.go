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

func NewContext(requestCtx *fasthttp.RequestCtx, logger logging.Loggerer) *Context {
	return &Context{
		Localized:  local.FromContext(requestCtx),
		Loggerer:   logger,
		requestCtx: requestCtx,
	}
}

func (self *Context) Request() *fasthttp.Request {
	return &self.requestCtx.Request
}

type Handler interface {
	Handle(ctx *Context) (int, error)
}
