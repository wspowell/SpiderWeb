package logging

import "github.com/sirupsen/logrus"

type Logger interface {
	Tag(name string, value interface{})
	Printf(format string, v ...interface{})
	Debug(format string, v ...interface{})
	Info(format string, v ...interface{})
	Warn(format string, v ...interface{})
	Error(format string, v ...interface{})
	Fatal(format string, v ...interface{})
}

type baseLog struct {
	logger *logrus.Logger

	cfg  Configer
	tags logrus.Fields
}

// Perf: Log value receivers performed better.
type Log struct {
	// Perf: Store least used values in one struct because copying one reference is faster than copying many values.
	*baseLog

	// Perf: Copying the level into the struct performed better.
	level Level
}

func NewLog(cfg Configer) Log {
	if logConfig, ok := cfg.(*Config); ok {
		return Log{
			baseLog: &baseLog{
				logger: logConfig.logger,
				cfg:    cfg,
				tags:   logrus.Fields{},
			},
			level: logConfig.level,
		}
	}

	return Log{
		baseLog: &baseLog{
			logger: newLogrusLogger(cfg),
			cfg:    cfg,
			tags:   logrus.Fields{},
		},
		level: cfg.Level(),
	}
}

func (self Log) Tag(name string, value interface{}) {
	self.tags[name] = value
}

func (self Log) Trace(format string, v ...interface{}) {
	if LevelTrace >= self.level {
		self.trace(format, v...)
	}
}

// Perf: Each log function has its own function based on benchmark feedback. Not sure why though.
func (self Log) trace(format string, v ...interface{}) {
	self.logger.WithFields(self.tags).WithFields(self.cfg.Tags()).Tracef(format, v...)
}

func (self Log) Debug(format string, v ...interface{}) {
	if LevelDebug >= self.level {
		self.debug(format, v...)
	}
}

func (self Log) debug(format string, v ...interface{}) {
	self.logger.WithFields(self.tags).WithFields(self.cfg.Tags()).Debugf(format, v...)
}

func (self Log) Printf(format string, v ...interface{}) {
	// Log at INFO to match logrus.
	if LevelInfo >= self.level {
		self.printf(format, v...)
	}
}

func (self Log) printf(format string, v ...interface{}) {
	self.logger.WithFields(self.tags).WithFields(self.cfg.Tags()).Printf(format, v...)
}

func (self Log) Info(format string, v ...interface{}) {
	if LevelInfo >= self.level {
		self.info(format, v...)
	}
}

func (self Log) info(format string, v ...interface{}) {
	self.logger.WithFields(self.tags).WithFields(self.cfg.Tags()).Infof(format, v...)
}

func (self Log) Warn(format string, v ...interface{}) {
	if LevelWarn >= self.level {
		self.warn(format, v...)
	}
}

func (self Log) warn(format string, v ...interface{}) {
	self.logger.WithFields(self.tags).WithFields(self.cfg.Tags()).Warnf(format, v...)
}

func (self Log) Error(format string, v ...interface{}) {
	if LevelError >= self.level {
		self.error(format, v...)
	}
}

func (self Log) error(format string, v ...interface{}) {
	self.logger.WithFields(self.tags).WithFields(self.cfg.Tags()).Errorf(format, v...)
}

func (self Log) Fatal(format string, v ...interface{}) {
	if LevelFatal >= self.level {
		self.fatal(format, v...)
	}
}

func (self Log) fatal(format string, v ...interface{}) {
	self.logger.WithFields(self.tags).WithFields(self.cfg.Tags()).Fatalf(format, v...)
}

func (self Log) Panic(format string, v ...interface{}) {
	if LevelPanic >= self.level {
		self.panic(format, v...)
	}
}

func (self Log) panic(format string, v ...interface{}) {
	self.logger.WithFields(self.tags).WithFields(self.cfg.Tags()).Panicf(format, v...)
}
