package endpoint

// Handler is the hook into the request handler.
// Handler struct should contain all data that should be
//   populated and validated before any business logic is run.
type Handler interface {
	// Handle business logic.
	// While not explicitly prevented, this function should
	//   not touch the request or do any post processing of
	//   any request data.
	Handle(ctx *Context) (int, error)
}
