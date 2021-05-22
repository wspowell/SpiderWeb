package endpoint

import (
	gohttp "net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wspowell/context"
	"github.com/wspowell/spiderweb/http"
)

var (
	uncachedHttpStatus = http.StatusOK
	cachedHttpStatus   = http.StatusNotModified
	uncachedResponse   = []byte("response not cached")
	cachedResponse     = []byte(nil)
	uncachedETag       = "uncached"
	cachedETag         = "19-2d477ab8aa9777a2f0c0275d17bd7647"
)

func Test_handleETag(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		description             string
		clientHeaders           map[string]string
		maxAgeSeconds           int
		httpStatus              int
		expectedResponseHeaders map[string]string
		expectedHttpStatus      int
		expectedResponseBody    []byte
	}{
		{
			description:             "non-success, no cache",
			clientHeaders:           map[string]string{},
			maxAgeSeconds:           0,
			httpStatus:              http.StatusBadRequest,
			expectedResponseHeaders: map[string]string{},
			expectedHttpStatus:      http.StatusBadRequest,
			expectedResponseBody:    uncachedResponse,
		},
		{
			description:             "no client headers, no max age, no etag",
			clientHeaders:           map[string]string{},
			maxAgeSeconds:           0,
			httpStatus:              uncachedHttpStatus,
			expectedResponseHeaders: map[string]string{},
			expectedHttpStatus:      uncachedHttpStatus,
			expectedResponseBody:    uncachedResponse,
		},
		{
			description: "IfNoneMatch client header with fresh cache, no max age, returns new etag",
			clientHeaders: map[string]string{
				http.HeaderIfNoneMatch: cachedETag,
			},
			maxAgeSeconds: 0,
			httpStatus:    uncachedHttpStatus,
			expectedResponseHeaders: map[string]string{
				http.HeaderETag: cachedETag,
			},
			expectedHttpStatus:   cachedHttpStatus,
			expectedResponseBody: cachedResponse,
		},
		{
			description: "IfNoneMatch Cache-Control=no-cache client header with fresh cache, no max age, no etag",
			clientHeaders: map[string]string{
				http.HeaderIfNoneMatch:  cachedETag,
				http.HeaderCacheControl: "no-cache",
			},
			maxAgeSeconds:           0,
			httpStatus:              uncachedHttpStatus,
			expectedResponseHeaders: map[string]string{},
			expectedHttpStatus:      uncachedHttpStatus,
			expectedResponseBody:    uncachedResponse,
		},
		{
			description: "IfNoneMatch client header with stale cache, no max age, returns new etag",
			clientHeaders: map[string]string{
				http.HeaderIfNoneMatch: uncachedETag,
			},
			maxAgeSeconds: 0,
			httpStatus:    uncachedHttpStatus,
			expectedResponseHeaders: map[string]string{
				http.HeaderETag: cachedETag,
			},
			expectedHttpStatus:   uncachedHttpStatus,
			expectedResponseBody: uncachedResponse,
		},
		{
			description: "IfNoneMatch client header with fresh cache, max age 300, returns new etag",
			clientHeaders: map[string]string{
				http.HeaderIfNoneMatch: cachedETag,
			},
			maxAgeSeconds: 300,
			httpStatus:    uncachedHttpStatus,
			expectedResponseHeaders: map[string]string{
				http.HeaderETag:         cachedETag,
				http.HeaderCacheControl: "max-age=300",
			},
			expectedHttpStatus:   cachedHttpStatus,
			expectedResponseBody: cachedResponse,
		},
		{
			description: "IfNoneMatch Cache-Control=no-cache client header with fresh cache, max age 300, returns new etag",
			clientHeaders: map[string]string{
				http.HeaderIfNoneMatch:  cachedETag,
				http.HeaderCacheControl: "no-cache",
			},
			maxAgeSeconds:           300,
			httpStatus:              uncachedHttpStatus,
			expectedResponseHeaders: map[string]string{},
			expectedHttpStatus:      uncachedHttpStatus,
			expectedResponseBody:    uncachedResponse,
		},
		{
			description: "IfNoneMatch client header with stale cache, max age 300, returns new etag",
			clientHeaders: map[string]string{
				http.HeaderIfNoneMatch: uncachedETag,
			},
			maxAgeSeconds: 300,
			httpStatus:    uncachedHttpStatus,
			expectedResponseHeaders: map[string]string{
				http.HeaderETag:         cachedETag,
				http.HeaderCacheControl: "max-age=300",
			},
			expectedHttpStatus:   uncachedHttpStatus,
			expectedResponseBody: uncachedResponse,
		},

		{
			description: "IfMatch client header with fresh cache, no max age",
			clientHeaders: map[string]string{
				http.HeaderIfMatch: cachedETag,
			},
			maxAgeSeconds: 0,
			httpStatus:    uncachedHttpStatus,
			expectedResponseHeaders: map[string]string{
				http.HeaderETag: cachedETag,
			},
			expectedHttpStatus:   uncachedHttpStatus,
			expectedResponseBody: uncachedResponse,
		},
		{
			description: "IfMatch Cache-Control=no-cache client header with fresh cache, no max age",
			clientHeaders: map[string]string{
				http.HeaderIfMatch:      cachedETag,
				http.HeaderCacheControl: "no-cache",
			},
			maxAgeSeconds:           0,
			httpStatus:              uncachedHttpStatus,
			expectedResponseHeaders: map[string]string{},
			expectedHttpStatus:      uncachedHttpStatus,
			expectedResponseBody:    uncachedResponse,
		},
		{
			description: "IfMatch client header with stale cache, no max age",
			clientHeaders: map[string]string{
				http.HeaderIfMatch: uncachedETag,
			},
			maxAgeSeconds: 0,
			httpStatus:    uncachedHttpStatus,
			expectedResponseHeaders: map[string]string{
				http.HeaderETag: cachedETag,
			},
			expectedHttpStatus:   http.StatusPreconditionFailed,
			expectedResponseBody: nil,
		},
		{
			description: "IfMatch client header with fresh cache, max age 300",
			clientHeaders: map[string]string{
				http.HeaderIfMatch: cachedETag,
			},
			maxAgeSeconds: 300,
			httpStatus:    uncachedHttpStatus,
			expectedResponseHeaders: map[string]string{
				http.HeaderETag:         cachedETag,
				http.HeaderCacheControl: "max-age=300",
			},
			expectedHttpStatus:   uncachedHttpStatus,
			expectedResponseBody: uncachedResponse,
		},
		{
			description: "IfMatch Cache-Control=no-cache client header with fresh cache, max age 300",
			clientHeaders: map[string]string{
				http.HeaderIfMatch:      cachedETag,
				http.HeaderCacheControl: "no-cache",
			},
			maxAgeSeconds:           300,
			httpStatus:              uncachedHttpStatus,
			expectedResponseHeaders: map[string]string{},
			expectedHttpStatus:      uncachedHttpStatus,
			expectedResponseBody:    uncachedResponse,
		},
		{
			description: "IfMatch client header with stale cache, max age 300",
			clientHeaders: map[string]string{
				http.HeaderIfMatch: uncachedETag,
			},
			maxAgeSeconds: 300,
			httpStatus:    uncachedHttpStatus,
			expectedResponseHeaders: map[string]string{
				http.HeaderETag:         cachedETag,
				http.HeaderCacheControl: "max-age=300",
			},
			expectedHttpStatus:   http.StatusPreconditionFailed,
			expectedResponseBody: nil,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			request, err := gohttp.NewRequest(http.MethodGet, "/", nil)
			assert.Nil(t, err)

			for key, value := range testCase.clientHeaders {
				request.Header[key] = []string{value}
			}

			requester := NewHttpRequester("/", request)

			httpStatus, responseBody := handleETag(ctx, requester, testCase.maxAgeSeconds, testCase.httpStatus, uncachedResponse)

			responseHeaders := requester.ResponseHeaders()
			for key, expectedValue := range testCase.expectedResponseHeaders {
				actualValue, exists := responseHeaders[key]
				assert.True(t, exists, "header does not exist: %v", key)
				assert.Equal(t, expectedValue, actualValue)
			}

			if httpStatus == http.StatusBadRequest {
				_, exists := responseHeaders[http.HeaderETag]
				assert.False(t, exists)
			}

			if testCase.maxAgeSeconds == 0 || httpStatus == http.StatusBadRequest {
				_, exists := responseHeaders[http.HeaderCacheControl]
				assert.False(t, exists)
			}

			assert.Equal(t, testCase.expectedHttpStatus, httpStatus)
			assert.Equal(t, testCase.expectedResponseBody, responseBody)
		})
	}
}
