// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"fmt"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/examples/hello"
	"github.com/mattermost/mattermost-plugin-apps/server/examples/hello/builtin_hello"
	"github.com/mattermost/mattermost-plugin-apps/server/examples/hello/http_hello"
	"github.com/mattermost/mattermost-plugin-apps/server/http/dialog"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

func (s *service) executeDebugClean(params *params) (*model.CommandResponse, error) {
	_ = s.api.Mattermost.KV.DeleteAll()
	return normalOut(params, md.MD("Deleted all KV records"), nil)
}

func (s *service) executeDebugBindings(params *params) (*model.CommandResponse, error) {
	bindings, err := s.api.Proxy.GetBindings(&api.Context{
		ActingUserID: params.commandArgs.UserId,
		UserID:       params.commandArgs.UserId,
		TeamID:       params.commandArgs.TeamId,
		ChannelID:    params.commandArgs.ChannelId,
	})
	if err != nil {
		return normalOut(params, md.MD("error"), err)
	}
	return normalOut(params, md.JSONBlock(bindings), nil)
}

func (s *service) executeDebugEmbeddedForm(params *params) (*model.CommandResponse, error) {
	cr := s.api.Proxy.Call(api.SessionToken(params.commandArgs.Session.Token), &api.Call{
		URL: hello.PathSendSurvey,
		Context: &api.Context{
			AppID:        http_hello.AppID,
			ActingUserID: params.commandArgs.UserId,
			ChannelID:    params.commandArgs.ChannelId,
			TeamID:       params.commandArgs.TeamId,
			UserID:       params.commandArgs.UserId,
		},
	})

	if cr.Type == api.CallResponseTypeError {
		return normalOut(params, md.MD("error"), cr)
	}

	return normalOut(params, md.MD("The app will send you the form"), nil)
}

func (s *service) executeDebugInstallHTTPHello(params *params) (*model.CommandResponse, error) {
	params.current = []string{
		"--app-secret", http_hello.AppSecret,
		"--url", s.api.Configurator.GetConfig().PluginURL + api.HelloHTTPPath + http_hello.PathManifest,
		"--force",
	}
	return s.executeInstall(params)
}

func (s *service) executeDebugInstallBuiltinHello(params *params) (*model.CommandResponse, error) {
	manifest := builtin_hello.GetManifest()

	app, _, err := s.api.Admin.ProvisionApp(
		&api.Context{
			ActingUserID: params.commandArgs.UserId,
		},
		api.SessionToken(params.commandArgs.Session.Token),
		&api.InProvisionApp{
			Manifest: manifest,
		},
	)
	if err != nil {
		return normalOut(params, nil, err)
	}

	conf := s.api.Configurator.GetConfig()

	// Finish the installation when the Dialog is submitted, see
	// <plugin>/http/dialog/install.go
	err = s.api.Mattermost.Frontend.OpenInteractiveDialog(
		dialog.NewInstallAppDialog(manifest, "", conf.PluginURL, params.commandArgs))
	if err != nil {
		return normalOut(params, nil, errors.Wrap(err, "couldn't open an interactive dialog"))
	}

	team, err := s.api.Mattermost.Team.Get(params.commandArgs.TeamId)
	if err != nil {
		return normalOut(params, nil, err)
	}

	return &model.CommandResponse{
		GotoLocation: params.commandArgs.SiteURL + "/" + team.Name + "/messages/@" + app.BotUsername,
		Text:         fmt.Sprintf("redirected to the DM with @%s to continue installing **%s**", app.BotUsername, manifest.DisplayName),
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
	}, nil
}
