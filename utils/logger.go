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
	Debugf(template string, args ...any)
	Debugw(msg string, keysAndValues ...any)
	Warnf(template string, args ...any)
	Warnw(msg string, keysAndValues ...any)
	Infof(template string, args ...any)
	Infow(msg string, keysAndValues ...any)
	Errorf(template string, args ...any)
	Errorw(msg string, keysAndValues ...any)
	Fatalf(template string, args ...any)
	Fatalw(msg string, keysAndValues ...any)

	// implemented here to provide a consistent interface, without using
	// *zap.SugaredLogger
	WithError(error) Logger
	With(args ...any) Logger
}

type NilLogger struct{}

var _ Logger = NilLogger{}

func (NilLogger) Debugf(string, ...any) {}
func (NilLogger) Debugw(string, ...any) {}
func (NilLogger) Warnf(string, ...any)  {}
func (NilLogger) Warnw(string, ...any)  {}
func (NilLogger) Infof(string, ...any)  {}
func (NilLogger) Infow(string, ...any)  {}
func (NilLogger) Errorf(string, ...any) {}
func (NilLogger) Errorw(string, ...any) {}
func (NilLogger) Fatalf(string, ...any) {}
func (NilLogger) Fatalw(string, ...any) {}

func (l NilLogger) WithError(error) Logger  { return l }
func (l NilLogger) With(args ...any) Logger { return l }

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
	Loggable() []any
}

// expandWith expands anything that implements LogProps into name, value pairs.
func expandWith(args []any) []any {
	var with []any

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

func (l *logger) With(args ...any) Logger {
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
	encodingConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

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

func LogDigest(i any) string {
	if s, ok := i.(string); ok {
		return s
	}

	var keys []string
	if m, ok := i.(map[string]any); ok {
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
