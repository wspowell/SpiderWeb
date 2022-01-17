package restful

import (
	"sync"

	"github.com/fasthttp/router"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"

	"github.com/wspowell/spiderweb/httpheader"
	"github.com/wspowell/spiderweb/httptrip"
)

var (
	bytesPool = sync.Pool{
		New: func() interface{} {
			return []byte{}
		},
	}
)

var _ httptrip.RoundTripper = (*fasthttpRequester)(nil)

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
	requestId := self.requestCtx.Request.Header.Peek(httpheader.XRequestId)
	if requestId == nil {
		return uuid.NewString() // This call is (relatively) slow. Would be nice to have a faster UUID package.
	}
	return string(requestId)
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

func (self *fasthttpRequester) PeekRequestHeader(key string) []byte {
	return self.requestCtx.Request.Header.Peek(key)
}

func (self *fasthttpRequester) VisitRequestHeaders(f func(key []byte, value []byte)) {
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
	// Note: Path is stored as a string, not []byte.
	//       See: fasthttp/router.go saveMatchedRoutePath()
	value, ok := self.requestCtx.UserValue(param).(string)

	return value, ok
}

func (self *fasthttpRequester) QueryParam(param string) (string, bool) {
	value := self.requestCtx.URI().QueryArgs().Peek(param)

	return string(value), value != nil
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

func (self *fasthttpRequester) ResponseContentType() []byte {
	return self.requestCtx.Response.Header.Peek(httpheader.ContentType)
}

func (self *fasthttpRequester) PeekResponseHeader(header string) []byte {
	return self.requestCtx.Response.Header.Peek(header)
}

func (self *fasthttpRequester) VisitResponseHeaders(f func(header []byte, value []byte)) {
	self.requestCtx.Response.Header.VisitAll(f)
}

func (self *fasthttpRequester) WriteResponse() {
	self.requestCtx.Response.SetStatusCode(self.statusCode)
	self.requestCtx.Response.SetBody(self.responseBodyBytes)
}
