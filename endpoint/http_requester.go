package endpoint

import (
	"io"
	"net/http"
	"strings"
)

type Requester interface {
	RequestId() string

	// HTTP Method.
	Method() []byte
	// Path of the actual request URL.
	Path() []byte

	ContentType() []byte
	Accept() []byte

	VisitHeaders(f func(key []byte, value []byte))

	// MatchedPath returns the endpoint path that the request URL matches.
	// Ex: /some/path/{id}
	MatchedPath() string

	// PathParam returns the path parameter value for the given parameter name.
	// Returns false if parameter not found.
	PathParam(param string) (string, bool)
	QueryParam(param string) ([]byte, bool)

	RequestBody() []byte

	SetResponseHeader(header string, value string)
	SetResponseContentType(contentType string)
	ResponseContentType() string
}

var _ Requester = (*HttpRequester)(nil)

type HttpRequester struct {
	matchedPath string
	request     *http.Request
	bodyBytes   []byte
}

func NewHttpRequester(matchedPath string, request *http.Request) *HttpRequester {
	var bodyBytes []byte
	if request.Body != nil {
		bodyBytes, _ = io.ReadAll(request.Body)
	}

	request.Response = &http.Response{
		Header: http.Header{},
	}

	return &HttpRequester{
		matchedPath: matchedPath,
		request:     request,
		bodyBytes:   bodyBytes,
	}
}

func (self *HttpRequester) RequestId() string {
	return "abc-123"
}

func (self *HttpRequester) Method() []byte {
	return []byte(self.request.Method)
}

func (self *HttpRequester) Path() []byte {
	return []byte(self.request.URL.Path)
}

func (self *HttpRequester) ContentType() []byte {
	return []byte(self.request.Header.Get("Content-Type"))
}

func (self *HttpRequester) Accept() []byte {
	return []byte(self.request.Header.Get("Accept"))
}

func (self *HttpRequester) VisitHeaders(f func(key []byte, value []byte)) {
	for header, value := range self.request.Header {
		f([]byte(header), []byte(value[0]))
	}
}

func (self *HttpRequester) MatchedPath() string {
	return self.matchedPath
}

func (self *HttpRequester) PathParam(param string) (string, bool) {
	urlParts := strings.Split(self.request.URL.Path, "/")
	pathParts := strings.Split(self.matchedPath, "/")

	for index, value := range pathParts {
		if value == "{"+param+"}" {
			return urlParts[index], true
		}
	}

	return "", false
}

func (self *HttpRequester) QueryParam(param string) ([]byte, bool) {
	value := self.request.URL.Query().Get(param)
	return []byte(value), value != ""
}

func (self *HttpRequester) RequestBody() []byte {
	return self.bodyBytes
}

func (self *HttpRequester) SetResponseHeader(header string, value string) {
	self.request.Response.Header.Set(header, value)
}

func (self *HttpRequester) SetResponseContentType(contentType string) {
	self.request.Response.Header.Set("Content-Type", contentType)
}

func (self *HttpRequester) ResponseContentType() string {
	return self.request.Response.Header.Get("Content-Type")
}
