package profiling_test

import (
	"testing"
	"time"

	"github.com/wspowell/context"

	"github.com/wspowell/spiderweb/profiling"
)

func Test_trace(t *testing.T) {
	t.Parallel()

	runProcess()
}

func runProcess() {
	ctx := context.Background()
	defer profiling.Profile(ctx, "runProcesses").Finish()

	timer := profiling.Profile(ctx, "manualDoOne")
	doOne(ctx)
	timer.Finish()

	timer = profiling.Profile(ctx, "manualDoTwo")
	doTwo(ctx)
	timer.Finish()

	time.Sleep(100 * time.Millisecond)
}

func doOne(ctx context.Context) {
	defer profiling.Profile(ctx, "doOne").Finish()
	time.Sleep(200 * time.Millisecond)
}

func doTwo(ctx context.Context) {
	defer profiling.Profile(ctx, "doTwo").Finish()
	time.Sleep(300 * time.Millisecond)
	doThree(ctx)
	time.Sleep(400 * time.Millisecond)
}

func doThree(ctx context.Context) {
	defer profiling.Profile(ctx, "doThree").Finish()
	time.Sleep(500 * time.Millisecond)
}
