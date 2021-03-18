// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"encoding/json"

	"github.com/mattermost/mattermost-server/v5/model"
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
	bindings, err := s.proxy.GetBindings(apps.SessionToken(params.commandArgs.Session.Token), apps.NewCommandContext(params.commandArgs))
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

	data, err := httputils.GetFromURL(manifestURL)
	if err != nil {
		return nil, err
	}
	m := apps.Manifest{}
	err = json.Unmarshal(data, &m)
	if err != nil {
		return errorOut(params, err)
	}

	out, err := s.proxy.AddLocalManifest(
		&apps.Context{ActingUserID: params.commandArgs.UserId},
		apps.SessionToken(params.commandArgs.Session.Token), &m)
	if err != nil {
		return errorOut(params, err)
	}
	return &model.CommandResponse{
		Text:         string(out),
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
	}, nil
}
