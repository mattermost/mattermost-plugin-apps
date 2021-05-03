// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

func (s *service) executeDebugClean(params *params) (*model.CommandResponse, error) {
	_ = s.mm.KV.DeleteAll()
	_ = s.conf.StoreConfig(config.StoredConfig{})
	return out(params, md.MD("Deleted all KV records and emptied the config."))
}

func (s *service) executeDebugBindings(params *params) (*model.CommandResponse, error) {
	bindings, err := s.proxy.GetBindings(
		params.commandArgs.Session.Id,
		params.commandArgs.UserId,
		s.newCommandContext(params.commandArgs))
	if err != nil {
		return errorOut(params, err)
	}
	return out(params, md.JSONBlock(bindings))
}

func (s *service) executeDebugAddManifest(params *params) (*model.CommandResponse, error) {
	manifestURL := ""
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	fs.StringVar(&manifestURL, "url", "", "manifest URL")
	if err := fs.Parse(params.current); err != nil {
		return errorOut(params, err)
	}

	if manifestURL == "" {
		return errorOut(params, errors.New("you must add a `--url`"))
	}

	data, err := httputils.GetFromURL(manifestURL)
	if err != nil {
		return errorOut(params, err)
	}

	m, err := apps.ManifestFromJSON(data)
	if err != nil {
		return errorOut(params, err)
	}

	out, err := s.proxy.AddLocalManifest(params.commandArgs.UserId, m)
	if err != nil {
		return errorOut(params, err)
	}
	return &model.CommandResponse{
		Text:         string(out),
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
	}, nil
}

func (s *service) newCommandContext(commandArgs *model.CommandArgs) *apps.Context {
	return s.conf.GetConfig().SetContextDefaults(&apps.Context{
		UserAgentContext: apps.UserAgentContext{
			TeamID:    commandArgs.TeamId,
			ChannelID: commandArgs.ChannelId,
		},
		ActingUserID: commandArgs.UserId,
		UserID:       commandArgs.UserId,
	})
}
