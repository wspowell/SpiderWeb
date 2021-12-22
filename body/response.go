package body

import "github.com/wspowell/spiderweb/mime"

type Responder interface {
	GetResponseBody() any
	MarshalResponseBody(mimeType mime.Handler) ([]byte, error)
}

type Response[T any] struct {
	ResponseBody T
}

func (self *Response[T]) GetResponseBody() any {
	return &self.ResponseBody
}

func (self *Response[T]) MarshalResponseBody(mimeType mime.Handler) ([]byte, error) {
	bodyBytes, err := mimeType.MarshalMimeType(self.GetResponseBody())
	if err != nil {
		return nil, err
	}
	return bodyBytes, nil
}
