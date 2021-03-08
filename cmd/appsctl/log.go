package main

import (
	"github.com/sirupsen/logrus"
)

var log logger

type logger struct {
	*logrus.Logger
}

func init() {
	log = logger{
		Logger: logrus.New(),
	}
}

func (l *logger) Error(message string, keyValuePairs ...interface{}) {
	l.WithFields(getFields(keyValuePairs...)).Error(message)
}

func (l *logger) Warn(message string, keyValuePairs ...interface{}) {
	l.WithFields(getFields(keyValuePairs...)).Warn(message)
}
func (l *logger) Info(message string, keyValuePairs ...interface{}) {
	l.WithFields(getFields(keyValuePairs...)).Info(message)
}
func (l *logger) Debug(message string, keyValuePairs ...interface{}) {
	l.WithFields(getFields(keyValuePairs...)).Debug(message)
}

func getFields(keyValuePairs ...interface{}) logrus.Fields {
	fields := logrus.Fields{}

	if len(keyValuePairs)%2 != 0 {
		panic("invalid logging key value data")
	}

	for i := 0; i < len(keyValuePairs); i += 2 {
		fields[keyValuePairs[i].(string)] = keyValuePairs[i+1]
	}

	return fields
}
