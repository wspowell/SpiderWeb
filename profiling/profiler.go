package profiling

import (
	"time"

	"github.com/wspowell/context"
)

type activeTimerKey struct{}

// Finisher interface for profiling timers.
type Finisher interface {
	// Finish the timer.
	Finish()
}

// Profile a given span.
// Must call Finish() on the returned timer.
func Profile(ctx context.Context, name string) Finisher {
	return newTiming(ctx, name)
}

type timing struct {
	ctx                 context.Context
	parent              *timing
	finishedChildTimers []*timing
	name                string
	start               time.Time
	duration            time.Duration
}

func newTiming(ctx context.Context, name string) *timing {
	var parentTimer *timing
	if activeTimer, ok := ctx.Value(activeTimerKey{}).(*timing); ok {
		parentTimer = activeTimer
	}

	timer := &timing{
		ctx:                 ctx,
		name:                name,
		parent:              parentTimer,
		finishedChildTimers: []*timing{},
		start:               time.Now().UTC(),
	}

	context.WithLocalValue(ctx, activeTimerKey{}, timer)

	return timer
}

func (self *timing) Finish() {
	self.duration = time.Since(self.start)
	if self.parent == nil {
		// Dump profiling data.
		// FIXME: This always prints profiling, even when LevelFatal is desired.
		// logConfig := log.NewConfig().WithLevel(log.LevelDebug)
		// logger := log.NewLogger(logConfig)
		// printTimers(logger, self, 0)
	} else {
		self.parent.finishedChildTimers = append(self.parent.finishedChildTimers, self)

		context.WithLocalValue(self.ctx, activeTimerKey{}, self.parent)
	}
}

// func printTimers(logger log.Logger, timer *timing, paddingSize int) {
// 	var padding string
// 	for i := 0; i < paddingSize; i++ {
// 		padding += " "
// 	}

// 	logger.Debug("%v%v -> %v", padding, timer.name, timer.duration)
// 	for _, childTimer := range timer.finishedChildTimers {
// 		printTimers(logger, childTimer, paddingSize+2)
// 	}

// }
