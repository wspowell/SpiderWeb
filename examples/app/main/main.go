package main

import (
	"github.com/wspowell/spiderweb/examples/app"
)

func main() {
	server := app.SetupServer()
	server.Listen()
}
