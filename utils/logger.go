// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package utils

import (
	"fmt"
	"sort"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const ErrorKey = "error"

type Logger interface {
	// from zap.SugaredLogger
	Debugf(template string, args ...interface{})
	Debugw(msg string, keysAndValues ...interface{})
	Warnf(template string, args ...interface{})
	Warnw(msg string, keysAndValues ...interface{})
	Infof(template string, args ...interface{})
	Infow(msg string, keysAndValues ...interface{})
	Errorf(template string, args ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
	Fatalf(template string, args ...interface{})
	Fatalw(msg string, keysAndValues ...interface{})

	// implemented here to provide a consistent interface, without using
	// *zap.SugaredLogger
	WithError(error) Logger
	With(args ...interface{}) Logger
}

type logger struct {
	*zap.SugaredLogger
}

func (l *logger) WithError(err error) Logger {
	if err == nil {
		return l
	}
	return &logger{
		SugaredLogger: l.SugaredLogger.With(ErrorKey, err.Error()),
	}
}

type HasLoggable interface {
	Loggable() []interface{}
}

// expandWith expands anything that implements LogProps into name, value pairs.
func expandWith(args []interface{}) []interface{} {
	var with []interface{}

	expectKeyOrProps := true
	for _, v := range args {
		lp, hasProps := v.(HasLoggable)
		_, isString := v.(string)
		switch {
		case !expectKeyOrProps:
			with = append(with, v)
			expectKeyOrProps = true
		case hasProps:
			with = append(with, expandWith(lp.Loggable())...)
		case !isString:
			with = append(with, "log_error", fmt.Sprintf("expected a string key or hasLogProps, found %T", v))
			return with
		default:
			// a string key.
			with = append(with, v)
			expectKeyOrProps = false
		}
	}
	return with
}

func (l *logger) With(args ...interface{}) Logger {
	return &logger{
		SugaredLogger: l.SugaredLogger.With(expandWith(args)...),
	}
}

func NewTestLogger() Logger {
	l, err := zap.NewDevelopmentConfig().Build()
	if err != nil {
		panic(err.Error())
	}
	return &logger{
		SugaredLogger: l.Sugar(),
	}
}

func MustMakeCommandLogger(level zapcore.Level) Logger {
	encodingConfig := zap.NewProductionEncoderConfig()
	encodingConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encodingConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encodingConfig.EncodeDuration = zapcore.StringDurationEncoder
	encodingConfig.EncodeCaller = zapcore.ShortCallerEncoder

	zconf := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Development:      false,
		Encoding:         "console",
		EncoderConfig:    encodingConfig,
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	l, err := zconf.Build()
	if err != nil {
		panic(err.Error())
	}
	return &logger{
		SugaredLogger: l.Sugar(),
	}
}

func LogDigest(i interface{}) string {
	if s, ok := i.(string); ok {
		return s
	}

	var keys []string
	if m, ok := i.(map[string]interface{}); ok {
		for key := range m {
			keys = append(keys, key)
		}
	}

	if m, ok := i.(map[string]string); ok {
		for key := range m {
			keys = append(keys, key)
		}
	}
	if len(keys) > 0 {
		sort.Strings(keys)
		return strings.Join(keys, ",")
	}

	return fmt.Sprintf("%v", i)
}
