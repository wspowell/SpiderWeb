package body

import "github.com/wspowell/spiderweb/mime"

type Requester interface {
	GetRequestBody() any
	UnmarshalRequestBody(bodyBytes []byte, mimeType mime.Handler) error
}

type Request[T any] struct {
	RequestBody T
}

func (self *Request[T]) GetRequestBody() any {
	return &self.RequestBody
}

func (self *Request[T]) UnmarshalRequestBody(bodyBytes []byte, mimeType mime.Handler) error {
	if err := mimeType.UnmarshalMimeType(bodyBytes, self.GetRequestBody()); err != nil {
		return err
	}
	return nil
}

type NoBody struct{}

func (self *NoBody) GetRequestBody() any {
	return nil
}

func (self *NoBody) UnmarshalRequestBody(bodyBytes []byte, mimeType mime.Handler) error {
	return nil
}
