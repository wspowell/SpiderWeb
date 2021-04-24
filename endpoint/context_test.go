package endpoint

import (
	"testing"
	"time"

	"github.com/wspowell/context"
	"github.com/wspowell/errors"
)

func Test_Context_DeadlineExceeded(t *testing.T) {
	t.Parallel()

	endpointCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	ctx := NewContext(endpointCtx, nil)

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

	endpointCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	ctx := NewContext(endpointCtx, nil)

	cancel()

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
