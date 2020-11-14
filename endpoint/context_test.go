package endpoint

import (
	"context"
	"testing"
	"time"

	"github.com/wspowell/spiderweb/errors"

	"github.com/valyala/fasthttp"
)

func Test_Context_DeadlineExceeded(t *testing.T) {
	t.Parallel()

	var req fasthttp.Request
	requestCtx := fasthttp.RequestCtx{}
	requestCtx.Init(&req, nil, nil)

	ctx := NewContext(context.Background(), &requestCtx, nil, time.Millisecond)

	time.Sleep(time.Second)

	if ctx.ShouldContinue() {
		t.Errorf("ShouldContinue() should have returned false")
	}

	if errors.Is(ctx.Context.Err(), context.Canceled) {
		t.Errorf("context error should not be Canceled")
	}

	if !errors.Is(ctx.Context.Err(), context.DeadlineExceeded) {
		t.Errorf("context error should be DeadlineExceeded")
	}
}

func Test_Context_Canceled(t *testing.T) {
	t.Parallel()

	var req fasthttp.Request
	requestCtx := fasthttp.RequestCtx{}
	requestCtx.Init(&req, nil, nil)

	ctx := NewContext(context.Background(), &requestCtx, nil, 30*time.Second)

	ctx.Cancel()

	if ctx.ShouldContinue() {
		t.Errorf("ShouldContinue() should have returned false")
	}

	if !errors.Is(ctx.Context.Err(), context.Canceled) {
		t.Errorf("context error should be Canceled")
	}

	if errors.Is(ctx.Context.Err(), context.DeadlineExceeded) {
		t.Errorf("context error should not be DeadlineExceeded")
	}
}
