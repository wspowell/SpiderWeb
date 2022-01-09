package body

import "github.com/wspowell/spiderweb/mime"

type Responder interface {
	GetResponseBody() any
	MarshalResponseBody(bodyBytes *[]byte, mimeType mime.Handler) error
}

type Response[T any] struct {
	ResponseBody T
}

func (self *Response[T]) GetResponseBody() any {
	return &self.ResponseBody
}

func (self *Response[T]) MarshalResponseBody(bodyBytes *[]byte, mimeType mime.Handler) error {
	return mimeType.MarshalMimeType(bodyBytes, self.GetResponseBody())
}
