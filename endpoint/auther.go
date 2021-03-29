package endpoint

// Auther defines request authentication.
type Auther interface {
	Auth(ctx *Context, VisitAllHeaders func(func(key, value []byte))) (int, error)
}
