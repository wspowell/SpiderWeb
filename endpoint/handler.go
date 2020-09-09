package endpoint

type Handler interface {
	Handle(ctx *Context) (int, error)
}
