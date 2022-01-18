package handler

import (
	"github.com/wspowell/context"

	"github.com/wspowell/errors"
	"github.com/wspowell/log"
	"github.com/wspowell/spiderweb/body"
	"github.com/wspowell/spiderweb/httpheader"
	"github.com/wspowell/spiderweb/httpstatus"
	"github.com/wspowell/spiderweb/httptrip"
	"github.com/wspowell/spiderweb/request"
	"github.com/wspowell/spiderweb/response"
)

type Runner struct {
	Handle
}

func (self *Runner) Run(ctx context.Context, reqRes httptrip.RoundTripper) {
	ctx, cancel := context.WithTimeout(ctx, self.timeout)
	defer cancel()

	// Every invocation creates its own logger instance.
	ctx = log.WithContext(ctx, self.logConfig)

	err := errors.Catch(func() {
		self.run(ctx, reqRes)
	})

	if err != nil {
		log.Error(ctx, "%s", err)
		log.Debug(ctx, "%+v", err)

		statusCode := httpstatus.InternalServerError
		responseBytes := reqRes.ResponseBody()
		self.errorResponse(&statusCode, &responseBytes, errors.Wrap(err, ErrInternalServerError))

		reqRes.SetStatusCode(statusCode)
		reqRes.SetResponseBody(responseBytes)
	}

	reqRes.WriteResponse()
}

func (self *Runner) run(ctx context.Context, reqRes httptrip.RoundTripper) {
	var err error

	// Check for timeout first. The request may have already waited too long.
	// This may occur if max concurrency is set too low or if the server has too many long running requests.
	if err = ctx.Err(); err != nil {
		statusCode := httpstatus.RequestTimeout
		responseBytes := reqRes.ResponseBody()
		self.errorResponse(&statusCode, &responseBytes, errors.Wrap(err, ErrTimeout))
		reqRes.SetStatusCode(statusCode)
		reqRes.SetResponseBody(responseBytes)
	}

	handlerInstance := self.newHandler()

	self.setLogTags(ctx, reqRes)
	reqRes.SetResponseContentType(string(reqRes.Accept())) // FIXME: should set a default for when not set on request.
	reqRes.SetResponseHeader(httpheader.XRequestId, reqRes.RequestId())

	if asRequester, ok := any(handlerInstance).(body.Requester); ok {
		if err = self.processRequest(reqRes, asRequester); err != nil {
			self.processError(reqRes, httpstatus.InternalServerError, err)
			return
		}
	}

	if err = self.processParameters(ctx, reqRes, handlerInstance); err != nil {
		self.processError(reqRes, httpstatus.InternalServerError, err)
		return
	}

	if asAuthorizer, ok := handlerInstance.(Authorizer); ok {
		if err := asAuthorizer.Authorize(reqRes); err != nil {
			self.processError(reqRes, httpstatus.Unauthorized, err)
			return
		}
	}

	if err = ctx.Err(); err != nil {
		self.processError(reqRes, httpstatus.RequestTimeout, errors.Wrap(err, ErrTimeout))
		return
	}

	// Run the endpoint
	var statusCode int
	if statusCode, err = handlerInstance.Handle(ctx); err != nil {
		self.processError(reqRes, statusCode, err)
		return
	}

	if err = ctx.Err(); err != nil {
		self.processError(reqRes, httpstatus.RequestTimeout, errors.Wrap(err, ErrTimeout))
		return
	}

	// Handle the response.
	if asResponder, ok := any(handlerInstance).(body.Responder); ok {
		if err = self.processResponse(reqRes, asResponder); err != nil {
			self.processError(reqRes, httpstatus.InternalServerError, err)
			return
		}
	}

	reqRes.SetStatusCode(statusCode)

	if self.maxAgeSeconds != 0 {
		log.Trace(ctx, "eTagEnabled, handling etag")
		response.HandleETag(ctx, reqRes, self.maxAgeSeconds, statusCode)
	}
}

func (self *Runner) setLogTags(ctx context.Context, reqRes httptrip.RoundTripper) {
	log.Tag(ctx, "request_id", reqRes.RequestId())
	log.Tag(ctx, "method", string(reqRes.Method()))
	log.Tag(ctx, "route", reqRes.MatchedPath())
	log.Tag(ctx, "path", string(reqRes.Path()))
	log.Tag(ctx, "action", self.action)
}

func (self *Runner) processRequest(reqRes httptrip.RoundTripper, req body.Requester) error {
	if mimeTypeHandler, exists := self.mimeTypes[string(reqRes.ContentType())]; exists {
		return req.UnmarshalRequestBody(reqRes.RequestBody(), mimeTypeHandler)
	}
	return errors.New("Content-Type mime type not supported: %s", reqRes.ContentType) // FIXME: wrap error
}

func (self *Runner) processResponse(reqRes httptrip.RoundTripper, res body.Responder) error {
	if mimeTypeHandler, exists := self.mimeTypes[string(reqRes.Accept())]; exists {
		responseBytes := reqRes.ResponseBody()
		err := res.MarshalResponseBody(&responseBytes, mimeTypeHandler)
		reqRes.SetResponseBody(responseBytes)
		return err
	}
	return errors.New("Accept mime type not supported: %s", reqRes.ContentType) // FIXME: wrap error
}

func (self *Runner) processParameters(ctx context.Context, reqRes httptrip.RoundTripper, e any) error {
	if asPathParams, ok := e.(request.Path); ok {
		for _, pathParam := range asPathParams.PathParameters() {
			if pathParamValue, exists := reqRes.PathParam(pathParam.ParamName()); exists {
				if err := pathParam.SetParam(pathParamValue); err != nil {
					return err // FIXME: wrap error
				}
				// Each path parameter is added as a log tag.
				// Note: It helps if the path parameter name is descriptive.
				log.Tag(ctx, string(pathParam.ParamName()), pathParamValue)
				continue
			}
			return errors.New("request does not have path parameter: %s", pathParam.ParamName()) // FIXME: wrap error
		}
	}

	if asQueryParams, ok := e.(request.Query); ok {
		for _, queryParam := range asQueryParams.QueryParameters() {
			if queryParamValue, exists := reqRes.QueryParam(queryParam.ParamName()); exists {
				if err := queryParam.SetParam(queryParamValue); err != nil {
					return err // FIXME: wrap error
				}
				continue
			}
			//return errors.New("request does not have query parameter: %s", queryParam.ParamName()) // FIXME: wrap error
		}
	}

	return nil
}

func (self *Runner) processError(reqRes httptrip.RoundTripper, statusCode int, err error) {
	responseBytes := reqRes.ResponseBody()
	self.errorResponse(&statusCode, &responseBytes, err)

	reqRes.SetStatusCode(statusCode)
	reqRes.SetResponseBody(responseBytes)
}
