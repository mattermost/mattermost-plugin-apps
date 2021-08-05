// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package utils

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
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

type plugin struct {
	zapcore.LevelEnabler
	plog   pluginapi.LogService
	fields map[string]zapcore.Field
}

func NewPluginLogger(client *pluginapi.Client) Logger {
	pc := &plugin{
		LevelEnabler: zapcore.DebugLevel,
		plog:         client.Log,
	}
	return &logger{
		SugaredLogger: zap.New(pc).Sugar(),
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

func (p *plugin) With(fields []zapcore.Field) zapcore.Core {
	return p.with(fields)
}

func (p *plugin) with(fields []zapcore.Field) *plugin {
	ff := map[string]zapcore.Field{}
	for k, v := range p.fields {
		ff[k] = v
	}
	for _, f := range fields {
		ff[f.Key] = f
	}
	c := *p
	c.fields = ff
	return &c
}

func (p *plugin) Check(e zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	return ce.AddCore(e, p)
}

func (p *plugin) Sync() error {
	return nil
}

func (p *plugin) Write(e zapcore.Entry, fields []zapcore.Field) error {
	p = p.with(fields)
	w := p.plog.Error
	switch e.Level {
	case zapcore.DebugLevel:
		w = p.plog.Debug
	case zapcore.InfoLevel:
		w = p.plog.Info
	case zapcore.WarnLevel:
		w = p.plog.Warn
	}

	pairs := []interface{}{}
	for k, f := range p.fields {
		switch {
		case f.Integer != 0:
			pairs = append(pairs, k, f.Integer)
		case f.String != "":
			pairs = append(pairs, k, f.String)
		default:
			pairs = append(pairs, k, f.Interface)
		}
	}
	w(e.Message, pairs...)
	return nil
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

func (l *logger) With(args ...interface{}) Logger {
	return &logger{
		SugaredLogger: l.SugaredLogger.With(args...),
	}
}
