package restful

import (
	"strconv"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"github.com/wspowell/spiderweb/http"
)

type fasthttpRequester struct {
	requestCtx *fasthttp.RequestCtx
}

func newFasthttpRequester(requestCtx *fasthttp.RequestCtx) *fasthttpRequester {
	return &fasthttpRequester{
		requestCtx: requestCtx,
	}
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
	return self.requestCtx.Request.Header.Peek(http.HeaderAccept)
}

func (self *fasthttpRequester) PeekHeader(key string) []byte {
	return self.requestCtx.Request.Header.Peek(string(key))
}

func (self *fasthttpRequester) VisitHeaders(f func(key []byte, value []byte)) {
	self.requestCtx.Request.Header.VisitAll(f)
}

func (self *fasthttpRequester) MatchedPath() string {
	var matchedPath string
	matchedPath, _ = self.requestCtx.UserValue(router.MatchedRoutePathParam).(string)
	return matchedPath
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
