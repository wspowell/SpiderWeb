package local_test

import (
	"context"
	"sync"
	"testing"

	"github.com/wspowell/spiderweb/local"
)

func checkContext(t *testing.T, ctx context.Context) {
	if ctx.Value("localKey") != nil {
		t.Errorf("expected 'localKey' to be nil")
	}

	if ctx.Value("duplicatedKey").(string) != "immutableValue" {
		t.Errorf("expected 'duplicatedKey' to be %v but was %v", "immutableValue", ctx.Value("duplicatedKey"))
	}

	if ctx.Value("immutable").(string) != "immutableValue" {
		t.Errorf("expected 'immutable' to be %v but was %v", "immutable", ctx.Value("immutable"))
	}
}

func checkLocal(t *testing.T, ctx local.Context) {
	if ctx.Value("localKey").(string) != "localValue" {
		t.Errorf("expected 'localKey' to be %v but was %v", "localValue", ctx.Value("localKey"))
	}

	if ctx.Value("duplicatedKey").(string) != "myValue" {
		t.Errorf("expected 'duplicatedKey' to be %v but was %v", "myValue", ctx.Value("duplicatedKey"))
	}

	if ctx.Value("immutable").(string) != "immutableValue" {
		t.Errorf("expected 'immutable' to be %v but was %v", "immutable", ctx.Value("immutable"))
	}
}

func Test_NewLocalized(t *testing.T) {
	t.Parallel()

	localCtx := local.NewLocalized()

	local.WithValue(localCtx, "immutable", "immutableValue")
	local.WithValue(localCtx, "duplicatedKey", "immutableValue")

	localCtx.Localize("localKey", "localValue")
	localCtx.Localize("duplicatedKey", "myValue")

	checkContext(t, localCtx.Context())
	checkLocal(t, localCtx)
}

func Test_FromContext(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(context.Background(), "immutable", "immutableValue")
	ctx = context.WithValue(ctx, "duplicatedKey", "immutableValue")

	localCtx := local.FromContext(ctx)
	localCtx.Localize("localKey", "localValue")
	localCtx.Localize("duplicatedKey", "myValue")

	checkContext(t, localCtx.Context())
	checkLocal(t, localCtx)
}

func Test_Context_ThreadSafety_Correct_Usage(t *testing.T) {
	t.Parallel()

	localCtx := local.NewLocalized()
	localCtx.Localize("localKey", "localValue")

	var paniced bool

	var wg sync.WaitGroup
	wg.Add(1)
	go func(ctx context.Context, wg *sync.WaitGroup) {
		defer func() {
			if err := recover(); err != nil {
				paniced = true
			}
		}()
		defer wg.Done()

		golocalCtx := local.FromContext(ctx)
		if golocalCtx.Value("localKey") == "localValue" {
			t.Errorf("local value should not be copied")
		}

		golocalCtx.Localize("localKey", "goroutineValue")

	}(localCtx, &wg)

	wg.Wait()

	if paniced {
		t.Errorf("unexpected panic")
	}

	if localCtx.Value("localKey") != "localValue" {
		t.Errorf("expected localKey == localValue")
	}
}
