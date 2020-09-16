package endpoint

// Auther defines request authentication.
type Auther interface {
	// TODO: #11 Pass in copies of the headers here instead of *fasthttp.Request.
	//       Consumers should not have to import fasthttp just for this.
	Auth(ctx *Context, headers map[string][]byte) (int, error)
}
