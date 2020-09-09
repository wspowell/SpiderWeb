package endpoint

import (
	"net/http"

	"github.com/valyala/fasthttp"
)

type Auther interface {
	// FIXME: It might be better to pass in copies of the headers here.
	Auth(request *fasthttp.Request) (int, error)
}

type Auth struct{}

func (self Auth) Auth(request *fasthttp.Request) (int, error) {
	return http.StatusOK, nil
}
