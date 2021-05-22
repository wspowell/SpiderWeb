package main

import (
	"github.com/wspowell/spiderweb/examples/restful/server"
)

func main() {
	custom := server.New()
	custom.Listen()
}
