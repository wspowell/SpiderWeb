package endpoint

import (
	"github.com/valyala/fasthttp"
)

type Auther interface {
	// TODO: Pass in copies of the headers here instead of *fasthttp.Request.
	//       Consumers should not have to import fasthttp just for this.
	Auth(request *fasthttp.Request) (int, error)
}
