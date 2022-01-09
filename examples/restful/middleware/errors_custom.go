package middleware

import (
	"github.com/wspowell/spiderweb/httpstatus"
)

func AllErrorsTeapot(statusCode *int, responseBody *[]byte, err error) {
	*statusCode = httpstatus.Teapot
}
