package route

import (
	"net/http"

	"github.com/wspowell/spiderweb/handler"
)

type Route struct {
	HttpMethod string
	Path       string
	Run        handler.Runner
}

func New(httpMethod string, path string, run handler.Runner) Route {
	return Route{
		HttpMethod: httpMethod,
		Path:       path,
		Run:        run,
	}
}

func Get(path string, run handler.Runner) Route {
	return Route{
		HttpMethod: http.MethodGet,
		Path:       path,
		Run:        run,
	}
}

func Post(path string, run handler.Runner) Route {
	return Route{
		HttpMethod: http.MethodPost,
		Path:       path,
		Run:        run,
	}
}

func Put(path string, run handler.Runner) Route {
	return Route{
		HttpMethod: http.MethodPut,
		Path:       path,
		Run:        run,
	}
}

func Patch(path string, run handler.Runner) Route {
	return Route{
		HttpMethod: http.MethodPatch,
		Path:       path,
		Run:        run,
	}
}

func Delete(path string, run handler.Runner) Route {
	return Route{
		HttpMethod: http.MethodDelete,
		Path:       path,
		Run:        run,
	}
}
