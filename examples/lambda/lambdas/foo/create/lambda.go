package main

import (
	"github.com/wspowell/spiderweb/handler"
	"github.com/wspowell/spiderweb/server/lambda"
)

func main() {
	handler := lambda.New("/foo", handler.NewHandle(create{}))
	handler.Start()
}
