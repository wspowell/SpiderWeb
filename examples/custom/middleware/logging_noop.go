package middleware

import (
	"io"

	"github.com/wspowell/logging"
)

type NoopLogConfig struct {
	*logging.Config
}

func (self *NoopLogConfig) Out() io.Writer {
	return io.Discard
}
