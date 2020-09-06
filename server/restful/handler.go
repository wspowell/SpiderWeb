package restful

import (
	"spiderweb/local"
)

type Handler interface {
	local.Localizer

	Handle() ([]byte, int)
}

type RoundTripper interface {
	RoundTrip(handler local.Localizer) ([]byte, int)
}

type Endpoint struct {
	Handler
}

func (self *Endpoint) RoundTrip(handler local.Localizer) ([]byte, int) {
	return self.Handler.Handle()
}
