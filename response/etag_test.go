package response_test

import (
	gohttp "net/http"
	"testing"
	"time"

	"github.com/wspowell/context"

	"github.com/stretchr/testify/assert"

	"github.com/wspowell/spiderweb/httpheader"
	"github.com/wspowell/spiderweb/httpmethod"
	"github.com/wspowell/spiderweb/httpstatus"
	"github.com/wspowell/spiderweb/httptrip"
	"github.com/wspowell/spiderweb/response"
)

const (
	uncachedHttpStatus     = httpstatus.OK
	cachedHttpStatus       = httpstatus.NotModified
	uncachedResponseString = "response not cached"
	uncachedETag           = "uncached"
	cachedETag             = "19-f563cf34dff2daac8d8e37fc17bd28ff60f79a05ed055116f82130ce136fab80"
)

func uncachedResponse() []byte {
	return []byte(uncachedResponseString)
}

func cachedResponse() []byte {
	return []byte(nil)
}

func Test_handleETag(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		description             string
		clientHeaders           map[string]string
		maxAgeSeconds           time.Duration
		httpStatus              int
		expectedResponseHeaders map[string]string
		expectedHttpStatus      int
		expectedResponseBody    []byte
	}{
		{
			description:             "non-success, no cache",
			clientHeaders:           map[string]string{},
			maxAgeSeconds:           0 * time.Second,
			httpStatus:              httpstatus.BadRequest,
			expectedResponseHeaders: map[string]string{},
			expectedHttpStatus:      httpstatus.BadRequest,
			expectedResponseBody:    uncachedResponse(),
		},
		{
			description:             "no client headers, no max age, no etag",
			clientHeaders:           map[string]string{},
			maxAgeSeconds:           0 * time.Second,
			httpStatus:              uncachedHttpStatus,
			expectedResponseHeaders: map[string]string{},
			expectedHttpStatus:      uncachedHttpStatus,
			expectedResponseBody:    uncachedResponse(),
		},
		{
			description: "IfNoneMatch client header with fresh cache, no max age, returns new etag",
			clientHeaders: map[string]string{
				httpheader.IfNoneMatch: cachedETag,
			},
			maxAgeSeconds: 0 * time.Second,
			httpStatus:    uncachedHttpStatus,
			expectedResponseHeaders: map[string]string{
				httpheader.ETag: cachedETag,
			},
			expectedHttpStatus:   cachedHttpStatus,
			expectedResponseBody: cachedResponse(),
		},
		{
			description: "IfNoneMatch Cache-Control=no-cache client header with fresh cache, no max age, no etag",
			clientHeaders: map[string]string{
				httpheader.IfNoneMatch:  cachedETag,
				httpheader.CacheControl: "no-cache",
			},
			maxAgeSeconds:           0 * time.Second,
			httpStatus:              uncachedHttpStatus,
			expectedResponseHeaders: map[string]string{},
			expectedHttpStatus:      uncachedHttpStatus,
			expectedResponseBody:    uncachedResponse(),
		},
		{
			description: "IfNoneMatch client header with stale cache, no max age, returns new etag",
			clientHeaders: map[string]string{
				httpheader.IfNoneMatch: uncachedETag,
			},
			maxAgeSeconds: 0 * time.Second,
			httpStatus:    uncachedHttpStatus,
			expectedResponseHeaders: map[string]string{
				httpheader.ETag: cachedETag,
			},
			expectedHttpStatus:   uncachedHttpStatus,
			expectedResponseBody: uncachedResponse(),
		},
		{
			description: "IfNoneMatch client header with fresh cache, max age 300, returns new etag",
			clientHeaders: map[string]string{
				httpheader.IfNoneMatch: cachedETag,
			},
			maxAgeSeconds: 300 * time.Second,
			httpStatus:    uncachedHttpStatus,
			expectedResponseHeaders: map[string]string{
				httpheader.ETag:         cachedETag,
				httpheader.CacheControl: "max-age=300",
			},
			expectedHttpStatus:   cachedHttpStatus,
			expectedResponseBody: cachedResponse(),
		},
		{
			description: "IfNoneMatch Cache-Control=no-cache client header with fresh cache, max age 300, returns new etag",
			clientHeaders: map[string]string{
				httpheader.IfNoneMatch:  cachedETag,
				httpheader.CacheControl: "no-cache",
			},
			maxAgeSeconds:           300 * time.Second,
			httpStatus:              uncachedHttpStatus,
			expectedResponseHeaders: map[string]string{},
			expectedHttpStatus:      uncachedHttpStatus,
			expectedResponseBody:    uncachedResponse(),
		},
		{
			description: "IfNoneMatch client header with stale cache, max age 300, returns new etag",
			clientHeaders: map[string]string{
				httpheader.IfNoneMatch: uncachedETag,
			},
			maxAgeSeconds: 300 * time.Second,
			httpStatus:    uncachedHttpStatus,
			expectedResponseHeaders: map[string]string{
				httpheader.ETag:         cachedETag,
				httpheader.CacheControl: "max-age=300",
			},
			expectedHttpStatus:   uncachedHttpStatus,
			expectedResponseBody: uncachedResponse(),
		},

		{
			description: "IfMatch client header with fresh cache, no max age",
			clientHeaders: map[string]string{
				httpheader.IfMatch: cachedETag,
			},
			maxAgeSeconds: 0 * time.Second,
			httpStatus:    uncachedHttpStatus,
			expectedResponseHeaders: map[string]string{
				httpheader.ETag: cachedETag,
			},
			expectedHttpStatus:   uncachedHttpStatus,
			expectedResponseBody: uncachedResponse(),
		},
		{
			description: "IfMatch Cache-Control=no-cache client header with fresh cache, no max age",
			clientHeaders: map[string]string{
				httpheader.IfMatch:      cachedETag,
				httpheader.CacheControl: "no-cache",
			},
			maxAgeSeconds:           0 * time.Second,
			httpStatus:              uncachedHttpStatus,
			expectedResponseHeaders: map[string]string{},
			expectedHttpStatus:      uncachedHttpStatus,
			expectedResponseBody:    uncachedResponse(),
		},
		{
			description: "IfMatch client header with stale cache, no max age",
			clientHeaders: map[string]string{
				httpheader.IfMatch: uncachedETag,
			},
			maxAgeSeconds: 0 * time.Second,
			httpStatus:    uncachedHttpStatus,
			expectedResponseHeaders: map[string]string{
				httpheader.ETag: cachedETag,
			},
			expectedHttpStatus:   httpstatus.PreconditionFailed,
			expectedResponseBody: nil,
		},
		{
			description: "IfMatch client header with fresh cache, max age 300",
			clientHeaders: map[string]string{
				httpheader.IfMatch: cachedETag,
			},
			maxAgeSeconds: 300 * time.Second,
			httpStatus:    uncachedHttpStatus,
			expectedResponseHeaders: map[string]string{
				httpheader.ETag:         cachedETag,
				httpheader.CacheControl: "max-age=300",
			},
			expectedHttpStatus:   uncachedHttpStatus,
			expectedResponseBody: uncachedResponse(),
		},
		{
			description: "IfMatch Cache-Control=no-cache client header with fresh cache, max age 300",
			clientHeaders: map[string]string{
				httpheader.IfMatch:      cachedETag,
				httpheader.CacheControl: "no-cache",
			},
			maxAgeSeconds:           300 * time.Second,
			httpStatus:              uncachedHttpStatus,
			expectedResponseHeaders: map[string]string{},
			expectedHttpStatus:      uncachedHttpStatus,
			expectedResponseBody:    uncachedResponse(),
		},
		{
			description: "IfMatch client header with stale cache, max age 300",
			clientHeaders: map[string]string{
				httpheader.IfMatch: uncachedETag,
			},
			maxAgeSeconds: 300 * time.Second,
			httpStatus:    uncachedHttpStatus,
			expectedResponseHeaders: map[string]string{
				httpheader.ETag:         cachedETag,
				httpheader.CacheControl: "max-age=300",
			},
			expectedHttpStatus:   httpstatus.PreconditionFailed,
			expectedResponseBody: nil,
		},
	}
	for index := range testCases {
		testCase := testCases[index]
		t.Run(testCase.description, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			req, err := gohttp.NewRequestWithContext(ctx, httpmethod.Get, "/", nil)
			assert.Nil(t, err)

			for key, value := range testCase.clientHeaders {
				req.Header[key] = []string{value}
			}

			reqRes, err := httptrip.NewHttpRoundTrip("/", req)
			assert.Nil(t, err)
			defer reqRes.Close()

			reqRes.SetStatusCode(testCase.httpStatus)
			reqRes.SetResponseBody(uncachedResponse())

			response.HandleETag(ctx, reqRes, testCase.maxAgeSeconds, testCase.httpStatus)

			responseHeaders := map[string]string{}
			reqRes.VisitResponseHeaders(func(header []byte, value []byte) {
				responseHeaders[string(header)] = string(value)
			})
			for key, expectedValue := range testCase.expectedResponseHeaders {
				actualValue, exists := responseHeaders[key]
				assert.True(t, exists, "header does not exist: %v", key)
				assert.Equal(t, expectedValue, actualValue)
			}

			if reqRes.StatusCode() == httpstatus.BadRequest {
				_, exists := responseHeaders[httpheader.ETag]
				assert.False(t, exists)
			}

			if testCase.maxAgeSeconds == 0 || reqRes.StatusCode() == httpstatus.BadRequest {
				_, exists := responseHeaders[httpheader.CacheControl]
				assert.False(t, exists)
			}

			assert.Equal(t, testCase.expectedHttpStatus, reqRes.StatusCode())
			assert.Equal(t, testCase.expectedResponseBody, reqRes.ResponseBody())
		})
	}
}
