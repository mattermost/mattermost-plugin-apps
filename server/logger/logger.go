package logger

import (
	apilogger "github.com/mattermost/mattermost-plugin-api/experimental/bot/logger"
)

type Logger = apilogger.Logger
type LogContext = apilogger.LogContext

type LogAPI interface {
	Error(message string, keyValuePairs ...interface{})
	Warn(message string, keyValuePairs ...interface{})
	Info(message string, keyValuePairs ...interface{})
	Debug(message string, keyValuePairs ...interface{})
}

type logWrapper struct {
	api LogAPI
}

func (l *logWrapper) LogError(message string, keyValuePairs ...interface{}) {
	l.api.Error(message, keyValuePairs)
}

func (l *logWrapper) LogWarn(message string, keyValuePairs ...interface{}) {
	l.api.Warn(message, keyValuePairs)
}

func (l *logWrapper) LogInfo(message string, keyValuePairs ...interface{}) {
	l.api.Info(message, keyValuePairs)
}

func (l *logWrapper) LogDebug(message string, keyValuePairs ...interface{}) {
	l.api.Debug(message, keyValuePairs)
}

func New(api LogAPI) Logger {
	lw := &logWrapper{api: api}

	return apilogger.New(lw)
}
