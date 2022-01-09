package response

import (
	"bytes"

	"github.com/wspowell/context"

	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/wspowell/log"

	"github.com/wspowell/spiderweb/httpheader"
	"github.com/wspowell/spiderweb/httptrip"
)

const (
	noCache  = "no-cache"
	comma    = ","
	anything = "*"
)

// handleETag passes through the http status and response if the cache is stale (or does not yet exist).
// If the cache is fresh and a success case with non-empty body, this will return 304 Not Modified with an empty body.
func HandleETag(ctx context.Context, reqRes httptrip.RoundTripper, maxAgeSeconds time.Duration, httpStatus int) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "handleETag()")
	defer span.Finish()

	ifNoneMatch := reqRes.PeekHeader(httpheader.IfNoneMatch)
	ifMatch := reqRes.PeekHeader(httpheader.IfMatch)
	cacheControl := reqRes.PeekHeader(httpheader.CacheControl)

	responseBody := reqRes.ResponseBody()

	// Simply return the current http status and response body if any:
	//   1. Not a success response (2xx)
	//   2. Response body is empty
	//   3. Request header Cache-Control is "no-cache"
	//   4. Neither header is set: If-None-Match, If-Match
	if !(httpStatus >= 200 && httpStatus < 300) ||
		len(responseBody) == 0 ||
		bytes.Contains(cacheControl, []byte(noCache)) ||
		(len(ifNoneMatch) == 0 && len(ifMatch) == 0) {
		log.Trace(ctx, "skipping etag check: httpStatus = %v, response body size = %v, Cache-Control = %v", httpStatus, len(responseBody), cacheControl)

		return
	}

	md5Sum := sha256.Sum256(responseBody)
	eTagValue := strconv.Itoa(len(responseBody)) + "-" + hex.EncodeToString(md5Sum[:])

	reqRes.SetResponseHeader(httpheader.ETag, eTagValue)
	if maxAgeSeconds != 0 {
		log.Trace(ctx, "etag max age seconds: %v", maxAgeSeconds)
		reqRes.SetResponseHeader(httpheader.CacheControl, "max-age="+strconv.Itoa(int(maxAgeSeconds.Seconds())))
	} else {
		log.Trace(ctx, "etag max age: indefinite")
	}

	if newHttpStatus, ok := isCacheFresh(ifNoneMatch, ifMatch, []byte(eTagValue)); ok {
		log.Trace(ctx, "etag fresh, not modified: %v", eTagValue)

		reqRes.SetStatusCode(newHttpStatus)
		reqRes.SetResponseBody(nil)
		return
	}

	log.Trace(ctx, "refreshed etag: %v", eTagValue)
}

// isCacheFresh check whether cache can be used in this HTTP request
func isCacheFresh(ifNoneMatch []byte, ifMatch []byte, eTagValue []byte) (int, bool) {
	if len(ifNoneMatch) != 0 {
		// Check for cache freshness.
		// Header If-None-Match
		return http.StatusNotModified, checkEtagNoneMatch(trimTags(bytes.Split(ifNoneMatch, []byte(comma))), eTagValue)
	}
	// Check etag precondition.
	// Header If-Match
	return http.StatusPreconditionFailed, checkEtagMatch(trimTags(bytes.Split(ifMatch, []byte(comma))), eTagValue)
}

func trimTags(tags [][]byte) [][]byte {
	trimedTags := make([][]byte, len(tags))

	for index, tag := range tags {
		trimedTags[index] = bytes.TrimSpace(tag)
	}

	return trimedTags
}

func checkEtagNoneMatch(etagsToNoneMatch [][]byte, eTagValue []byte) bool {
	for _, etagToNoneMatch := range etagsToNoneMatch {
		if bytes.Equal(etagToNoneMatch, []byte(anything)) || bytes.Equal(etagToNoneMatch, eTagValue) {
			return true
		}
	}

	return false
}

func checkEtagMatch(etagsToMatch [][]byte, eTagValue []byte) bool {
	for _, etagToMatch := range etagsToMatch {
		if bytes.Equal(etagToMatch, []byte(anything)) {
			return false
		}

		if bytes.Equal(etagToMatch, eTagValue) {
			return false
		}
	}

	return true
}
