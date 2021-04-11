package main

import (
	"github.com/wspowell/spiderweb/examples/custom/server"
)

func main() {
	custom := server.New()
	custom.Listen()
}
