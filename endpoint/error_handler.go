package endpoint

type ErrorHandler interface {
	HandleError(ctx *Context, httpStatus int, err error) (int, []byte)
}
