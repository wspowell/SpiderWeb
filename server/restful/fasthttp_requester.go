package restful

import (
	"strconv"
	"sync"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"

	"github.com/wspowell/spiderweb/httpheader"
)

var (
	bytesPool = sync.Pool{
		New: func() interface{} {
			// The Pool's New function should generally only return pointer
			// types, since a pointer can be put into the return interface
			// value without an allocation:
			return []byte{}
		},
	}
)

type fasthttpRequester struct {
	requestCtx        *fasthttp.RequestCtx
	statusCode        int
	responseBodyBytes []byte
}

func newFasthttpRequester(requestCtx *fasthttp.RequestCtx) *fasthttpRequester {
	responseBodyBytes := bytesPool.Get().([]byte)
	responseBodyBytes = responseBodyBytes[:0]

	return &fasthttpRequester{
		requestCtx:        requestCtx,
		responseBodyBytes: responseBodyBytes,
	}
}

func (self *fasthttpRequester) Close() {
	bytesPool.Put(self.responseBodyBytes)
}

func (self *fasthttpRequester) RequestId() string {
	return strconv.Itoa(int(self.requestCtx.ID()))
}

func (self *fasthttpRequester) Method() []byte {
	return self.requestCtx.Method()
}

func (self *fasthttpRequester) Path() []byte {
	return self.requestCtx.URI().Path()
}

func (self *fasthttpRequester) ContentType() []byte {
	return self.requestCtx.Request.Header.ContentType()
}

func (self *fasthttpRequester) Accept() []byte {
	return self.requestCtx.Request.Header.Peek(httpheader.Accept)
}

func (self *fasthttpRequester) PeekHeader(key string) []byte {
	return self.requestCtx.Request.Header.Peek(key)
}

func (self *fasthttpRequester) VisitHeaders(f func(key []byte, value []byte)) {
	self.requestCtx.Request.Header.VisitAll(f)
}

func matchedPath(requestCtx *fasthttp.RequestCtx) string {
	var matchedPath string
	matchedPath, _ = requestCtx.UserValue(router.MatchedRoutePathParam).(string)

	return matchedPath
}

func (self *fasthttpRequester) MatchedPath() string {
	return matchedPath(self.requestCtx)
}

func (self *fasthttpRequester) PathParam(param string) (string, bool) {
	value, ok := self.requestCtx.UserValue(param).(string)

	return value, ok
}

func (self *fasthttpRequester) QueryParam(param string) ([]byte, bool) {
	value := self.requestCtx.URI().QueryArgs().Peek(param)

	return value, value != nil
}

func (self *fasthttpRequester) RequestBody() []byte {
	return self.requestCtx.Request.Body()
}

func (self *fasthttpRequester) ResponseBody() []byte {
	return self.responseBodyBytes
}

func (self *fasthttpRequester) StatusCode() int {
	return self.statusCode
}

func (self *fasthttpRequester) SetStatusCode(statusCode int) {
	self.statusCode = statusCode
}

func (self *fasthttpRequester) SetResponseBody(body []byte) {
	self.responseBodyBytes = body
}

func (self *fasthttpRequester) SetResponseHeader(header string, value string) {
	self.requestCtx.Response.Header.Set(header, value)
}

func (self *fasthttpRequester) SetResponseContentType(contentType string) {
	self.requestCtx.SetContentType(contentType)
}

func (self *fasthttpRequester) ResponseContentType() string {
	return string(self.requestCtx.Response.Header.Peek("Content-Type"))
}

func (self *fasthttpRequester) ResponseHeaders() map[string]string {
	headers := map[string]string{}
	self.requestCtx.Response.Header.VisitAll(func(key []byte, value []byte) {
		headers[string(key)] = string(value)
	})

	return headers
}

func (self *fasthttpRequester) WriteResponse() {
	self.requestCtx.Response.SetStatusCode(self.statusCode)
	self.requestCtx.Response.SetBody(self.responseBodyBytes)
}
