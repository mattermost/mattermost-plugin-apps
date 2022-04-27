// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package utils

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
)

// plugin implements zapcore.Core interface to serve as a logging "backend" for
// SugaredLogger, using the plugin API log methods.
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
