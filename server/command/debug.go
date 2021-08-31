// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func (s *service) executeDebugClean(params *commandParams) (*model.CommandResponse, error) {
	_ = s.conf.MattermostAPI().KV.DeleteAll()
	_ = s.conf.StoreConfig(config.StoredConfig{})
	return out(params, "Deleted all KV records and emptied the config.")
}

func (s *service) executeDebugBindings(params *commandParams) (*model.CommandResponse, error) {
	bindings, err := s.proxy.GetBindings(
		params.commandArgs.Session.Id,
		params.commandArgs.UserId,
		s.newCommandContext(params.commandArgs))
	if err != nil {
		return errorOut(params, err)
	}
	return out(params, utils.JSONBlock(bindings))
}

func (s *service) executeDebugAddManifest(params *commandParams) (*model.CommandResponse, error) {
	manifestURL := ""
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	fs.StringVar(&manifestURL, "url", "", "manifest URL")
	if err := fs.Parse(params.current); err != nil {
		return errorOut(params, err)
	}

	if manifestURL == "" {
		return errorOut(params, errors.New("you must add a `--url`"))
	}

	// Inside a debug command: all URLs are trusted.
	data, err := s.httpOut.GetFromURL(manifestURL, true, apps.MaxManifestSize)
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
		Text:         out,
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
	}, nil
}
