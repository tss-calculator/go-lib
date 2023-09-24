package logger

import (
	"time"

	"github.com/tss-calculator/go-lib/pkg/application/logger"

	"github.com/sirupsen/logrus"
)

const appNameKey = "app_name"

type Config struct {
	AppName string
}

func NewLogger(config *Config) logger.MainLogger {
	impl := logrus.New()
	impl.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
		FieldMap:        fieldMap,
	})

	return &loggerImpl{
		FieldLogger: impl.WithField(appNameKey, config.AppName),
	}
}

type loggerImpl struct {
	logrus.FieldLogger
}

func (l *loggerImpl) WithField(key string, value interface{}) logger.Logger {
	return &loggerImpl{l.FieldLogger.WithField(key, value)}
}

func (l *loggerImpl) WithFields(fields logger.Fields) logger.Logger {
	return &loggerImpl{l.FieldLogger.WithFields(logrus.Fields(fields))}
}

func (l *loggerImpl) Error(err error, args ...interface{}) {
	l.FieldLogger.WithError(err).Error(args...)
}

func (l *loggerImpl) FatalError(err error, args ...interface{}) {
	l.FieldLogger.WithError(err).Fatal(args...)
}

var fieldMap = logrus.FieldMap{
	logrus.FieldKeyTime: "@timestamp",
	logrus.FieldKeyMsg:  "message",
}
