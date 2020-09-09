package auth

import (
	"net/http"

	"github.com/valyala/fasthttp"
)

type Noop struct{}

func (self Noop) Auth(request *fasthttp.Request) (int, error) {
	return http.StatusOK, nil
}
