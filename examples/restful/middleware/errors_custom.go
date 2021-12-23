package middleware

import (
	"github.com/wspowell/spiderweb/httpstatus"
)

func AllErrorsTeapot(statusCode int, err error) (int, []byte) {
	return httpstatus.Teapot, nil
}
