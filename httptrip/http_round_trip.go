package httptrip

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/wspowell/errors"
)

var (
	ErrInvalidBody = errors.New("invalid body")
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
	return "abc-123"
}

func (self *HttpRoundTrip) Method() []byte {
	return []byte(self.request.Method)
}

func (self *HttpRoundTrip) Path() []byte {
	return []byte(self.request.URL.Path)
}

func (self *HttpRoundTrip) ContentType() []byte {
	return []byte(self.request.Header.Get("Content-Type"))
}

func (self *HttpRoundTrip) Accept() []byte {
	return []byte(self.request.Header.Get("Accept"))
}

func (self *HttpRoundTrip) PeekHeader(key string) []byte {
	if value, exists := self.request.Header[key]; exists {
		return []byte(value[0])
	}

	return nil
}

func (self *HttpRoundTrip) VisitHeaders(f func(key []byte, value []byte)) {
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

	for index, value := range pathParts {
		if value == "{"+param+"}" {
			return urlParts[index], true
		}
	}

	return "", false
}

func (self *HttpRoundTrip) QueryParam(param string) ([]byte, bool) {
	value := self.request.URL.Query().Get(param)

	return []byte(value), value != ""
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
	self.request.Response.Header.Set("Content-Type", contentType)
}

func (self *HttpRoundTrip) ResponseContentType() string {
	return self.request.Response.Header.Get("Content-Type")
}

func (self *HttpRoundTrip) ResponseHeaders() map[string]string {
	headers := map[string]string{}
	for key, value := range self.request.Response.Header {
		headers[key] = value[0]
	}

	return headers
}

func (self *HttpRoundTrip) WriteResponse() {
	self.request.Response.Body = ioutil.NopCloser(bytes.NewBuffer(self.responseBodyBytes))
	self.request.Response.StatusCode = self.statusCode
}
