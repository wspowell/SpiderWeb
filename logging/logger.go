package logging

import (
	"github.com/sirupsen/logrus"
)

type Loggerer interface {
	LogConfig() Configurer
	Tag(name string, value interface{})
	Printf(format string, v ...interface{})
	Debug(format string, v ...interface{})
	Info(format string, v ...interface{})
	Warn(format string, v ...interface{})
	Error(format string, v ...interface{})
	Fatal(format string, v ...interface{})
}

func copyTags(tags map[string]interface{}) map[string]interface{} {
	tagsCopy := make(map[string]interface{}, len(tags))
	for name, value := range tags {
		tagsCopy[name] = value
	}
	return tagsCopy
}

type Logger struct {
	config Configurer
	logger *logrus.Logger
	tags   logrus.Fields
}

func NewLogger(config Configurer) *Logger {
	logger := logrus.New()
	logger.Out = config.Out()

	if logrusLevel, err := logrus.ParseLevel(config.Level().String()); err != nil {
		logger.Fatalf("invalid logger level: %v", config.Level().String())
	} else {
		logger.Level = logrusLevel
	}

	tags := config.GlobalTags()

	return &Logger{
		config: config,
		logger: logger,
		tags:   tags,
	}
}

func (self *Logger) LogConfig() Configurer {
	return self.config.Clone()
}

func (self *Logger) Tag(name string, value interface{}) {
	self.tags[name] = value
}

func (self *Logger) Printf(format string, v ...interface{}) {
	self.logger.WithFields(self.tags).Printf(format, v...)
}

func (self *Logger) Debug(format string, v ...interface{}) {
	self.logger.WithFields(self.tags).Debugf(format, v...)
}

func (self *Logger) Info(format string, v ...interface{}) {
	self.logger.WithFields(self.tags).Infof(format, v...)
}

func (self *Logger) Warn(format string, v ...interface{}) {
	self.logger.WithFields(self.tags).Warnf(format, v...)
}

func (self *Logger) Error(format string, v ...interface{}) {
	self.logger.WithFields(self.tags).Errorf(format, v...)
}

func (self *Logger) Fatal(format string, v ...interface{}) {
	self.logger.WithFields(self.tags).Fatalf(format, v...)
}

func (self *Logger) SetLogger(loggerConfig Configurer) {
	logger := NewLogger(loggerConfig)
	*self = *logger
}
