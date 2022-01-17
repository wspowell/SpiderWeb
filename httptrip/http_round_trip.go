package httptrip

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/wspowell/errors"
	"github.com/wspowell/spiderweb/httpheader"
)

var (
	ErrInvalidBody = errors.New("invalid body")
)

var (
	bytesPool = sync.Pool{
		New: func() interface{} {
			return []byte{}
		},
	}
)

var _ RoundTripper = (*HttpRoundTrip)(nil)

type HttpRoundTrip struct {
	matchedPath       string
	request           *http.Request
	requestBodyBytes  []byte
	statusCode        int
	responseBodyBytes []byte
}

func NewHttpRoundTrip(matchedPath string, request *http.Request) (*HttpRoundTrip, error) {
	var requestBodyBytes []byte
	if request.Body != nil {
		b := bytesPool.Get().([]byte)
		b = b[:0]
		bufWriter := bytes.NewBuffer(b)

		if _, err := io.Copy(bufWriter, request.Body); err != nil {
			bytesPool.Put(b)
			return nil, errors.Wrap(err, ErrInvalidBody)
		}

		requestBodyBytes = bufWriter.Bytes()
	}

	responseBodyBytes := bytesPool.Get().([]byte)
	responseBodyBytes = responseBodyBytes[:0]

	request.Response = &http.Response{
		Header: http.Header{},
	}

	return &HttpRoundTrip{
		matchedPath:       matchedPath,
		request:           request,
		requestBodyBytes:  requestBodyBytes,
		responseBodyBytes: responseBodyBytes,
	}, nil
}

func (self *HttpRoundTrip) Close() {
	bytesPool.Put(self.requestBodyBytes)
	bytesPool.Put(self.responseBodyBytes)
}

func (self *HttpRoundTrip) RequestId() string {
	requestId, exists := self.request.Header[httpheader.XRequestId]
	if !exists {
		return uuid.New().String()
	}
	return requestId[0]
}

func (self *HttpRoundTrip) Method() []byte {
	return []byte(self.request.Method)
}

func (self *HttpRoundTrip) Path() []byte {
	return []byte(self.request.URL.Path)
}

func (self *HttpRoundTrip) ContentType() []byte {
	return []byte(self.request.Header.Get(httpheader.ContentType))
}

func (self *HttpRoundTrip) Accept() []byte {
	return []byte(self.request.Header.Get(httpheader.Accept))
}

func (self *HttpRoundTrip) PeekRequestHeader(key string) []byte {
	if value, exists := self.request.Header[key]; exists {
		return []byte(value[0])
	}

	return nil
}

func (self *HttpRoundTrip) VisitRequestHeaders(f func(key []byte, value []byte)) {
	for header, value := range self.request.Header {
		f([]byte(header), []byte(value[0]))
	}
}

func (self *HttpRoundTrip) MatchedPath() string {
	return self.matchedPath
}

func (self *HttpRoundTrip) PathParam(param string) (string, bool) {
	urlParts := strings.Split(self.request.URL.Path, "/")
	pathParts := strings.Split(self.matchedPath, "/")
	paramVariable := "{" + param + "}"

	for index, value := range pathParts {
		if value == paramVariable {
			return urlParts[index], true
		}
	}

	return "", false
}

func (self *HttpRoundTrip) QueryParam(param string) (string, bool) {
	value := self.request.URL.Query().Get(string(param))

	return value, value != ""
}

func (self *HttpRoundTrip) RequestBody() []byte {
	return self.requestBodyBytes
}

func (self *HttpRoundTrip) ResponseBody() []byte {
	return self.responseBodyBytes
}

func (self *HttpRoundTrip) StatusCode() int {
	return self.statusCode
}

func (self *HttpRoundTrip) SetStatusCode(statusCode int) {
	self.statusCode = statusCode
}

func (self *HttpRoundTrip) SetResponseHeader(header string, value string) {
	self.request.Response.Header.Set(header, value)
}

func (self *HttpRoundTrip) SetResponseBody(body []byte) {
	self.responseBodyBytes = body
}

func (self *HttpRoundTrip) SetResponseContentType(contentType string) {
	self.request.Response.Header.Set(httpheader.ContentType, contentType)
}

func (self *HttpRoundTrip) ResponseContentType() []byte {
	return []byte(self.request.Response.Header.Get(httpheader.ContentType))
}

func (self *HttpRoundTrip) PeekResponseHeader(header string) []byte {
	value, exists := self.request.Response.Header[header]
	if exists {
		return []byte(value[0])
	}
	return nil
}

func (self *HttpRoundTrip) VisitResponseHeaders(f func(header []byte, value []byte)) {
	for key, value := range self.request.Response.Header {
		f([]byte(key), []byte(value[0]))
	}
}

func (self *HttpRoundTrip) WriteResponse() {
	self.request.Response.Body = ioutil.NopCloser(bytes.NewBuffer(self.responseBodyBytes))
	self.request.Response.StatusCode = self.statusCode
}
