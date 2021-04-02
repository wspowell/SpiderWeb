package spiderwebtest

import (
	"bytes"
	"os"
	"runtime/debug"
	"testing"

	"github.com/wspowell/spiderweb"

	fuzz "github.com/google/gofuzz"
	"github.com/valyala/fasthttp"
)

type requestTestCase struct {
	httpMethod      string
	path            string
	requestMimeType string
	requestBody     []byte

	httpStatus       int
	responseMimeType string
	responseBody     []byte
}

// GivenRequest starts a request test case to be provided to TestRequest.
func GivenRequest(httpMethod string, path string) *requestTestCase {
	return &requestTestCase{
		httpMethod: httpMethod,
		path:       path,
	}
}

// WithRequestBody sets a request body for the request test case.
// This is optional.
func (self *requestTestCase) WithRequestBody(mimeType string, requestBody []byte) *requestTestCase {
	self.requestMimeType = mimeType
	self.requestBody = requestBody
	return self
}

// Expect the response to match the given status and body.
func (self *requestTestCase) Expect(httpStatus int, mimeType string, responseBody []byte) *requestTestCase {
	self.httpStatus = httpStatus
	self.responseMimeType = mimeType
	self.responseBody = responseBody
	return self
}

// TestRequest for request/response roundtrip.
func TestRequest(t *testing.T, server *spiderweb.Server, testCase *requestTestCase) {
	copyRequestBody := make([]byte, len(testCase.requestBody))
	copyResponseBody := make([]byte, len(testCase.responseBody))

	copy(copyRequestBody, testCase.requestBody)
	copy(copyResponseBody, testCase.responseBody)

	copyRequestTestCase := requestTestCase{
		httpMethod:       testCase.httpMethod,
		path:             testCase.path,
		requestMimeType:  testCase.requestMimeType,
		requestBody:      copyRequestBody,
		httpStatus:       testCase.httpStatus,
		responseMimeType: testCase.responseMimeType,
		responseBody:     copyResponseBody,
	}

	var req fasthttp.Request

	req.Header.SetMethod(copyRequestTestCase.httpMethod)
	req.Header.SetRequestURI(copyRequestTestCase.path)
	req.Header.Set(fasthttp.HeaderHost, "localhost")
	req.Header.Set("Content-Type", copyRequestTestCase.requestMimeType)
	req.Header.Set("Accept", copyRequestTestCase.responseMimeType)
	req.SetBody(copyRequestTestCase.requestBody)

	requestCtx := fasthttp.RequestCtx{}
	requestCtx.Init(&req, nil, nil)

	actualHttpStatus, actualResponseBody := server.Execute(&requestCtx)

	if copyRequestTestCase.httpStatus != actualHttpStatus {
		t.Errorf("expected http status %v, but got %v", copyRequestTestCase.httpStatus, actualHttpStatus)
	}

	if !bytes.Equal(copyRequestTestCase.responseBody, actualResponseBody) {
		t.Errorf("expected request body '%v', but got '%v'", string(copyRequestTestCase.responseBody), string(actualResponseBody))
	}

	requestFuzzTest(t, server, testCase.httpMethod, testCase.path)
}

func requestFuzzTest(t *testing.T, server *spiderweb.Server, httpMethod string, path string) {
	if doFuzz, exists := os.LookupEnv("FUZZ"); !exists || doFuzz != "true" {
		return
	}

	var requestBody []byte
	defer func() {
		if err := recover(); err != nil {
			t.Fatalf("%+v\route: %v %v\nrequest body: %+v\n%+v", err, httpMethod, path, string(requestBody), string(debug.Stack()))
		}
	}()

	f := fuzz.New()

	for i := 0; i < 100; i++ {
		f.Fuzz(&requestBody)

		var req fasthttp.Request

		req.Header.SetMethod(httpMethod)
		req.Header.SetRequestURI(path)
		req.Header.Set(fasthttp.HeaderHost, "localhost")
		req.SetBody(requestBody)

		requestCtx := fasthttp.RequestCtx{}
		requestCtx.Init(&req, nil, nil)

		server.Execute(&requestCtx)
	}
}
