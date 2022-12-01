// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package utils

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v6/model"
)

type LogConfig struct {
	BotUserID   string
	ChannelID   string
	Level       zapcore.Level
	IncludeJSON bool
}

type LogConfigGetter interface {
	GetLogConfig() LogConfig
}

// plugin implements zapcore.Core interface to serve as a logging "backend" for
// SugaredLogger, using the plugin API log methods.
type plugin struct {
	zapcore.LevelEnabler
	poster     *pluginapi.PostService
	logger     *pluginapi.LogService
	confGetter LogConfigGetter
	fields     map[string]zapcore.Field
}

func NewPluginLogger(mmapi *pluginapi.Client, confGetter LogConfigGetter) Logger {
	return &logger{
		SugaredLogger: zap.New(&plugin{
			poster:       &mmapi.Post,
			logger:       &mmapi.Log,
			LevelEnabler: zapcore.DebugLevel,
			confGetter:   confGetter,
		}).Sugar(),
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
	w := p.logger.Error
	switch e.Level {
	case zapcore.DebugLevel:
		w = p.logger.Debug
	case zapcore.InfoLevel:
		w = p.logger.Info
	case zapcore.WarnLevel:
		w = p.logger.Warn
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

	if p.confGetter == nil {
		return nil
	}
	logconf := p.confGetter.GetLogConfig()

	if logconf.ChannelID == "" || !logconf.Level.Enabled(e.Level) {
		return nil
	}

	message := fmt.Sprintf("%s %s: %s", e.Time.Format(time.StampMilli), e.Level.CapitalString(), e.Message)

	if logconf.IncludeJSON {
		ccJSON := map[string]any{}
		for i := 0; i < len(pairs); i += 2 {
			ccJSON[pairs[i].(string)] = pairs[i+1]
		}
		if len(ccJSON) > 0 {
			message += JSONBlock(ccJSON)
		}
	}

	logPost := &model.Post{
		ChannelId: logconf.ChannelID,
		UserId:    logconf.BotUserID,
		Message:   message,
	}
	if err := p.poster.CreatePost(logPost); err != nil {
		// Log directly to avoid loop
		p.logger.Error("failed to post log message", "err", err.Error())
	}

	return nil
}
