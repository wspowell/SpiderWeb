package main

import (
	"github.com/wspowell/spiderweb/examples/restful/resources/db"
	"github.com/wspowell/spiderweb/examples/restful/server"
)

func main() {
	database := db.NewDatabase()

	custom := server.New(database)
	custom.Listen()
}
