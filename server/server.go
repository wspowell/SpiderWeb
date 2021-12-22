package server

import "github.com/wspowell/spiderweb/mime"

type Server struct {
	MimeTypes map[string]mime.Handler
}
