package endpoint

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"strconv"

	"github.com/wspowell/context"
	"github.com/wspowell/log"
	"github.com/wspowell/spiderweb/http"
)

var (
	noCache = []byte("no-cache")
	comma   = []byte(",")
	any     = []byte("*")
)

// handleETag passes through the http status and response if the cache is stale (or does not yet exist).
// If the cache is fresh and a success case with non-empty body, this will return 304 Not Modified with an empty body.
func handleETag(ctx context.Context, requester Requester, maxAgeSeconds int, httpStatus int, responseBody []byte) (int, []byte) {
	ifNoneMatch := requester.PeekHeader(http.HeaderIfNoneMatch)
	ifMatch := requester.PeekHeader(http.HeaderIfMatch)
	cacheControl := requester.PeekHeader(http.HeaderCacheControl)

	// Simply return the current http status and response body if any:
	//   1. Not a success response (2xx)
	//   2. Response body is empty
	//   3. Request header Cache-Control is "no-cache"
	//   4. Neither header is set: If-None-Match, If-Match
	if !(200 <= httpStatus && httpStatus < 300) ||
		len(responseBody) == 0 ||
		bytes.Contains(cacheControl, noCache) ||
		(len(ifNoneMatch) == 0 && len(ifMatch) == 0) {
		log.Debug(ctx, "skipping etag check: httpStatus = %v, response body size = %v, Cache-Control = %v", httpStatus, len(responseBody), cacheControl)
		return httpStatus, responseBody
	}

	md5Sum := md5.Sum(responseBody)
	eTagValue := strconv.Itoa(len(responseBody)) + "-" + hex.EncodeToString(md5Sum[:])

	requester.SetResponseHeader(http.HeaderETag, eTagValue)
	if maxAgeSeconds != 0 {
		log.Debug(ctx, "etag max age seconds: %v", maxAgeSeconds)
		requester.SetResponseHeader(http.HeaderCacheControl, "max-age="+strconv.Itoa(maxAgeSeconds))
	} else {
		log.Debug(ctx, "etag max age: indefinite")
	}

	if newHttpStatus, ok := isCacheFresh(requester, ifNoneMatch, ifMatch, []byte(eTagValue)); ok {
		log.Debug(ctx, "etag fresh, not modified: %v", eTagValue)
		return newHttpStatus, nil
	}
	log.Debug(ctx, "refreshed etag: %v", eTagValue)
	return httpStatus, responseBody
}

// isCacheFresh check whether cache can be used in this HTTP request
func isCacheFresh(requester Requester, ifNoneMatch []byte, ifMatch []byte, eTagValue []byte) (int, bool) {
	if len(ifNoneMatch) != 0 {
		// Check for cache freshness.
		// Header If-None-Match
		return http.StatusNotModified, checkEtagNoneMatch(trimTags(bytes.Split(ifNoneMatch, comma)), eTagValue)
	}
	// Check etag precondition.
	// Header If-Match
	return http.StatusPreconditionFailed, checkEtagMatch(trimTags(bytes.Split(ifMatch, comma)), eTagValue)
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
		if bytes.Equal(etagToNoneMatch, any) || bytes.Equal(etagToNoneMatch, eTagValue) {
			return true
		}
	}

	return false
}

func checkEtagMatch(etagsToMatch [][]byte, eTagValue []byte) bool {
	for _, etagToMatch := range etagsToMatch {
		if bytes.Equal(etagToMatch, any) {
			return false
		}

		if bytes.Equal(etagToMatch, eTagValue) {
			return false
		}
	}

	return true
}
