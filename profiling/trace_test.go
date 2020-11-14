package profiling

import (
	"testing"
	"time"

	"github.com/wspowell/local"
)

func Test_trace(t *testing.T) {
	t.Parallel()

	runProcess()
}

func runProcess() {
	ctx := local.NewLocalized()
	defer Profile(ctx, "runProcesses").Finish()

	timer := Profile(ctx, "manualDoOne")
	doOne(ctx)
	timer.Finish()

	timer = Profile(ctx, "manualDoTwo")
	doTwo(ctx)
	timer.Finish()

	time.Sleep(100 * time.Millisecond)
}

func doOne(ctx local.Context) {
	defer Profile(ctx, "doOne").Finish()
	time.Sleep(200 * time.Millisecond)
}

func doTwo(ctx local.Context) {
	defer Profile(ctx, "doTwo").Finish()
	time.Sleep(300 * time.Millisecond)
	doThree(ctx)
	time.Sleep(400 * time.Millisecond)
}

func doThree(ctx local.Context) {
	defer Profile(ctx, "doThree").Finish()
	time.Sleep(500 * time.Millisecond)
}
