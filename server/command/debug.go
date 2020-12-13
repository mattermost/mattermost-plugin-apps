// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"fmt"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/examples/aws_hello"
	"github.com/mattermost/mattermost-plugin-apps/server/examples/hello"
	"github.com/mattermost/mattermost-plugin-apps/server/examples/hello/builtin_hello"
	"github.com/mattermost/mattermost-plugin-apps/server/examples/hello/http_hello"
	"github.com/mattermost/mattermost-plugin-apps/server/http/dialog"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

func (s *service) executeDebugClean(params *params) (*model.CommandResponse, error) {
	_ = s.api.Mattermost.KV.DeleteAll()
	_ = s.api.Configurator.StoreConfig(&api.StoredConfig{})
	return normalOut(params, md.MD("Deleted all KV records and emptied the config."), nil)
}

func (s *service) executeDebugBindings(params *params) (*model.CommandResponse, error) {
	bindings, err := s.api.Proxy.GetBindings(api.NewCommandContext(params.commandArgs))
	if err != nil {
		return normalOut(params, md.MD("error"), err)
	}
	return normalOut(params, md.JSONBlock(bindings), nil)
}

func (s *service) executeDebugEmbeddedForm(params *params) (*model.CommandResponse, error) {
	cc := api.NewCommandContext(params.commandArgs)
	cc.AppID = http_hello.AppID

	cr := s.api.Proxy.Call(
		api.SessionToken(params.commandArgs.Session.Token),
		&api.Call{
			URL:     hello.PathSendSurvey,
			Context: cc,
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
	manifest := builtin_hello.Manifest()

	app, _, err := s.api.Admin.ProvisionApp(
		&api.Context{
			ActingUserID: params.commandArgs.UserId,
		},
		api.SessionToken(params.commandArgs.Session.Token),
		&api.InProvisionApp{
			Manifest: manifest,
			Force:    true,
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

func (s *service) executeDebugInstallAWSHello(params *params) (*model.CommandResponse, error) {
	manifest := aws_hello.Manifest()

	app, _, err := s.api.Admin.ProvisionApp(
		&api.Context{
			ActingUserID: params.commandArgs.UserId,
		},
		api.SessionToken(params.commandArgs.Session.Token),
		&api.InProvisionApp{
			Manifest: manifest,
			Force:    true,
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
		Text:         fmt.Sprintf("%s. redirected to the DM with @%s to continue installing **%s**", "text", app.BotUsername, manifest.DisplayName),
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
	}, nil
}
