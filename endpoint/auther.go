package endpoint

import (
	"net/http"
)

type Auther interface {
	Auth(request *http.Request) (int, error)
}

type Auth struct{}

func (self Auth) Auth(request *http.Request) (int, error) {
	return http.StatusOK, nil
}
