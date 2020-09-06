package restful

import (
	"net/http"
)

type Server interface {
	ServeRestful(request *http.Request, writer http.ResponseWriter)
}

type Router interface {
	Route(httpMethod string, path string, handler Handler)
}
