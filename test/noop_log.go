package test

import (
	"io"

	"github.com/wspowell/log"
)

type NoopLogConfig struct {
	log.Config
}

func (self *NoopLogConfig) Out() io.Writer {
	return io.Discard
}
