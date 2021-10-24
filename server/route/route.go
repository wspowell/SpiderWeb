package route

import (
	"net/http"

	"github.com/wspowell/spiderweb/endpoint"
)

type Route struct {
	HttpMethod string
	Path       string
	Handler    endpoint.Handler
}

func New(httpMethod string, path string, handler endpoint.Handler) Route {
	return Route{
		HttpMethod: httpMethod,
		Path:       path,
		Handler:    handler,
	}
}

func Get(path string, handler endpoint.Handler) Route {
	return Route{
		HttpMethod: http.MethodGet,
		Path:       path,
		Handler:    handler,
	}
}

func Post(path string, handler endpoint.Handler) Route {
	return Route{
		HttpMethod: http.MethodPost,
		Path:       path,
		Handler:    handler,
	}
}

func Put(path string, handler endpoint.Handler) Route {
	return Route{
		HttpMethod: http.MethodPut,
		Path:       path,
		Handler:    handler,
	}
}

func Patch(path string, handler endpoint.Handler) Route {
	return Route{
		HttpMethod: http.MethodPatch,
		Path:       path,
		Handler:    handler,
	}
}

func Delete(path string, handler endpoint.Handler) Route {
	return Route{
		HttpMethod: http.MethodDelete,
		Path:       path,
		Handler:    handler,
	}
}
