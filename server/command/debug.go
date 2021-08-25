// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/spf13/pflag"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func (s *service) executeDebugClean(params *commandParams) *model.CommandResponse {
	_ = s.conf.MattermostAPI().KV.DeleteAll()
	_ = s.conf.StoreConfig(config.StoredConfig{})
	return out(params, "Deleted all KV records and emptied the config.")
}

func (s *service) executeDebugBindings(params *commandParams) *model.CommandResponse {
	bindings, err := s.proxy.GetBindings(
		params.commandArgs.Session.Id,
		params.commandArgs.UserId,
		s.newCommandContext(params.commandArgs))
	if err != nil {
		return s.errorOut(params, nil, err)
	}
	return out(params, utils.JSONBlock(bindings))
}

func (s *service) executeDebugAddManifest(params *commandParams) *model.CommandResponse {
	manifestURL := ""
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	fs.StringVar(&manifestURL, "url", "", "manifest URL")
	if err := fs.Parse(params.current); err != nil {
		return s.errorOut(params, nil, err)
	}

	if manifestURL == "" {
		return s.errorOut(params, utils.NewLocError(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "apps.command.debug.addManifest.error.url",
				Other: "you must add a `--url`",
			},
		}), nil)
	}

	// Inside a debug command: all URLs are trusted.
	data, err := s.httpOut.GetFromURL(manifestURL, true)
	if err != nil {
		return s.errorOut(params, nil, err)
	}

	m, err := apps.ManifestFromJSON(data)
	if err != nil {
		return s.errorOut(params, nil, err)
	}

	out, err := s.proxy.AddLocalManifest(params.commandArgs.UserId, m)
	if err != nil {
		return s.errorOut(params, nil, err)
	}

	return &model.CommandResponse{
		Text:         out,
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
	}
}
