package endpoint

import (
	"net/http"

	"spiderweb/local"
	"spiderweb/logging"
)

type Context struct {
	local.Context
	logging.Loggerer

	request *http.Request
}

func NewContext(request *http.Request, logger logging.Loggerer) *Context {
	return &Context{
		Context:  local.NewLocalized(request.Context()),
		Loggerer: logger,
		request:  request,
	}
}

func (self *Context) Request() *http.Request {
	return self.request
}

type Handler interface {
	Handle(ctx *Context) (int, error)
}
