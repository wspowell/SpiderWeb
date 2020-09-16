package endpoint

type ErrorHandler interface {
	MimeType() string
	HandleError(ctx *Context, httpStatus int, err error) (int, []byte)
}
