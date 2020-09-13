package main

import (
	"spiderweb/examples/app"
)

func main() {
	server := app.SetupServer()
	server.Listen()
}
