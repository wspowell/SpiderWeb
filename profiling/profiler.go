package profiling

import (
	"time"

	"github.com/wspowell/spiderweb/local"
	"github.com/wspowell/spiderweb/logging"
)

type contextKey struct{}

var activeTimerKey = contextKey{}

// Finisher interface for profiling timers.
type Finisher interface {
	// Finish the timer.
	Finish()
}

// Profile a given span.
// Must call Finish() on the returned timer.
func Profile(ctx local.Context, name string) Finisher {
	return newTiming(ctx, name)
}

type timing struct {
	ctx                 local.Context
	parent              *timing
	finishedChildTimers []*timing
	name                string
	start               time.Time
	duration            time.Duration
}

func newTiming(ctx local.Context, name string) *timing {
	var parentTimer *timing
	if activeTimer, ok := ctx.Value(activeTimerKey).(*timing); ok {
		parentTimer = activeTimer
	}

	timer := &timing{
		ctx:                 ctx,
		name:                name,
		parent:              parentTimer,
		finishedChildTimers: []*timing{},
		start:               time.Now().UTC(),
	}

	ctx.Localize(activeTimerKey, timer)

	return timer
}

func (self *timing) Finish() {
	self.duration = time.Since(self.start)
	if self.parent == nil {
		// Dump profiling data.
		logConfig := logging.NewConfig(logging.LevelDebug, map[string]interface{}{})
		logger := logging.NewLogger(logConfig)

		printTimers(logger, self, 0)
	} else {
		self.parent.finishedChildTimers = append(self.parent.finishedChildTimers, self)

		self.ctx.Localize(activeTimerKey, self.parent)
	}
}

func printTimers(logger logging.Loggerer, timer *timing, paddingSize int) {
	var padding string
	for i := 0; i < paddingSize; i++ {
		padding += " "
	}

	logger.Debug("%v%v -> %v", padding, timer.name, timer.duration)
	for _, childTimer := range timer.finishedChildTimers {
		printTimers(logger, childTimer, paddingSize+2)
	}

}
