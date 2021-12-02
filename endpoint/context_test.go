package endpoint_test

import (
	"testing"
	"time"

	"github.com/wspowell/context"
	"github.com/wspowell/errors"

	"github.com/wspowell/spiderweb/endpoint"
)

func Test_Context_DeadlineExceeded(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	time.Sleep(time.Second)

	if endpoint.ShouldContinue(ctx) {
		t.Errorf("ShouldContinue() should have returned false")
	}

	if errors.Is(ctx.Err(), context.Canceled) {
		t.Errorf("context error should not be Canceled")
	}

	if !errors.Is(ctx.Err(), context.DeadlineExceeded) {
		t.Errorf("context error should be DeadlineExceeded")
	}
}

func Test_Context_Canceled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	cancel()

	if endpoint.ShouldContinue(ctx) {
		t.Errorf("ShouldContinue() should have returned false")
	}

	if !errors.Is(ctx.Err(), context.Canceled) {
		t.Errorf("context error should be Canceled")
	}

	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		t.Errorf("context error should not be DeadlineExceeded")
	}
}
