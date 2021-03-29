package logging

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

type Level uint8

const (
	LevelTrace = Level(iota)
	LevelDebug = Level(iota)
	LevelInfo  = Level(iota)
	LevelWarn  = Level(iota)
	LevelError = Level(iota)
	LevelFatal = Level(iota)
	LevelPanic = Level(iota)
)

func (self Level) String() string {
	return [...]string{
		"trace",
		"debug",
		"info",
		"warn",
		"error",
		"fatal",
		"panic",
	}[self]
}

type Configer interface {
	Level() Level
	// Tags are added to each Logger created.
	// Therefore, these tags are global and must not be altered.
	Tags() map[string]interface{}
	Out() io.Writer
}

type Config struct {
	logger     *logrus.Logger
	level      Level
	globalTags map[string]interface{}
}

func NewConfig(level Level, globalTags map[string]interface{}) *Config {
	cfg := &Config{
		level:      level,
		globalTags: globalTags,
	}

	cfg.logger = newLogrusLogger(cfg)

	return cfg
}

func (self *Config) Level() Level {
	return self.level
}

func (self *Config) Tags() map[string]interface{} {
	return self.globalTags
}

func (self *Config) Out() io.Writer {
	return os.Stdout
}

func newLogrusLogger(cfg Configer) *logrus.Logger {
	logger := logrus.New()
	logger.Out = cfg.Out()

	logrusLevel, err := logrus.ParseLevel(cfg.Level().String())
	if err != nil {
		logger.Fatalf("invalid logger level: %v", cfg.Level().String())
	}

	logger.Level = logrusLevel

	return logger
}
