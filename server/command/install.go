// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/http/dialog"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"

	"github.com/mattermost/mattermost-plugin-apps/server/constants"
)

func (s *service) executeInstall(params *params) (*model.CommandResponse, error) {
	manifestURL := ""
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	fs.StringVar(&manifestURL, "url", "", "manifest URL")

	err := fs.Parse(params.current)
	if err != nil {
		return normalOut(params, nil, err)
	}

	var manifest apps.Manifest
	resp, err := http.Get(manifestURL)
	if err != nil {
		return normalOut(params, nil, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return normalOut(params, nil, errors.Errorf("expected 200 OK, got %v %v", resp.StatusCode, resp.Status))
	}
	err = json.NewDecoder(resp.Body).Decode(&manifest)
	if err != nil {
		return normalOut(params, nil, err)
	}

	conf := s.apps.Config.GetConfig()
	post := &model.Post{
		Message: fmt.Sprintf("Installing App: **%s**", manifest.DisplayName),
	}

	err = s.apps.Mattermost.Post.DM(conf.BotUserID, params.commandArgs.UserId, post)
	if err != nil {
		return normalOut(params, nil, err)
	}

	err = s.apps.Mattermost.Frontend.OpenInteractiveDialog(
		dialog.NewInstallAppDialog(
			params.commandArgs.TriggerId,
			&manifest,
			s.apps.Config.GetConfig().PluginURL,
			post.Id))
	if err != nil {
		return normalOut(params, nil, errors.Wrap(err, "couldn't open an interactive dialog"))
	}

	team, err := s.apps.Mattermost.Team.Get(params.commandArgs.TeamId)
	if err != nil {
		return normalOut(params, nil, err)
	}

	return &model.CommandResponse{
		GotoLocation: params.commandArgs.SiteURL + "/" + team.Name + "/messages/@" + constants.BotUserName,
		Text:         "redirected to the DM with @" + constants.BotUserName,
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
	}, nil
}
