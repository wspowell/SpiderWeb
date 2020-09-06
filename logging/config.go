package logging

import (
	"io"
	"os"
)

type Level int

const (
	LevelDebug = Level(iota)
	LevelInfo  = Level(iota)
	LevelWarn  = Level(iota)
	LevelError = Level(iota)
	LevelFatal = Level(iota)
)

func (self Level) String() string {
	return [...]string{
		"debug",
		"info",
		"warn",
		"error",
		"fatal",
	}[self]
}

type Configurer interface {
	Level() Level
	GlobalTags() map[string]interface{}
	Out() io.Writer
	Copy() Configurer
}

type Config struct {
	level      Level
	globalTags map[string]interface{}
}

func NewConfig(level Level, globalTags map[string]interface{}) *Config {
	return &Config{
		level:      level,
		globalTags: copyTags(globalTags),
	}
}

func (self *Config) Level() Level {
	return self.level
}

func (self *Config) GlobalTags() map[string]interface{} {
	return copyTags(self.globalTags)
}

func (self *Config) Out() io.Writer {
	return os.Stdout
}

func (self *Config) Copy() Configurer {
	return &Config{
		level:      self.level,
		globalTags: copyTags(self.globalTags),
	}
}
