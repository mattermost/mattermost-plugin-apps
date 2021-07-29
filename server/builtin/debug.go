// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func (a *builtinApp) debugCommandBinding() *apps.Binding {
	return &apps.Binding{
		Label:    "debug",
		Location: "debug",
		Bindings: []*apps.Binding{
			commandBinding("clean", pDebugClean, "", "remove all Apps and reset the persistent store"),
			commandBinding("bindings", pDebugBindings, "", "display all bindings for the current context"),
		},
	}
}

func (a *builtinApp) debugClean(creq *apps.CallRequest) *apps.CallResponse {
	// <>/<> !! TODO!!!
	return mdResponse("Deleted all KV records and emptied the config.")
}

// func (a *builtinApp) debugInstall(call *api.Call) (md.MD, error) {
// 	appID := api.AppID(call.GetStringValue(fieldExampleApp, ""))
// 	fmt.Printf("<><> debugInstall 1: appID: %s\n", appID)

// 	manifest := builtin_hello.Manifest()

// 	app, _, err := a.API.Admin.ProvisionApp(
// 		&api.Context{
// 			ActingUserID: params.commandArgs.UserId,
// 		},
// 		api.SessionToken(params.commandArgs.Session.Token),
// 		&api.InProvisionApp{
// 			Manifest: manifest,
// 			Force:    true,
// 		},
// 	)
// 	if err != nil {
// 		return errorOut(params, err)
// 	}

// 	conf := s.api.Configurator.GetConfig()

// 	// Finish the installation when the Dialog is submitted, see
// 	// <plugin>/http/dialog/install.go
// 	err = s.api.Mattermost.Frontend.OpenInteractiveDialog(
// 		dialog.NewInstallAppDialog(manifest, "", conf.PluginURL, params.commandArgs))
// 	if err != nil {
// 		return errorOut(params, errors.Wrap(err, "couldn't open an interactive dialog"))
// 	}

// 	team, err := s.api.Mattermost.Team.Get(params.commandArgs.TeamId)
// 	if err != nil {
// 		return errorOut(params, err)
// 	}

// 	return &model.CommandResponse{
// 		GotoLocation: params.commandArgs.SiteURL + "/" + team.Name + "/messages/@" + app.BotUsername,
// 		Text:         fmt.Sprintf("redirected to the DM with @%s to continue installing **%s**", app.BotUsername, manifest.DisplayName),
// 		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
// 	}, nil

// 	return "", nil
// }

// func (s *service) executeDebugInstallHTTPHello(params *params) (*model.CommandResponse, error) {
// 	params.current = []string{
// 		"--app-secret", http_hello.AppSecret,
// 		"--url", s.api.Configurator.GetConfig().PluginURL + api.HelloHTTPPath + http_hello.PathManifest,
// 		"--force",
// 	}
// 	return s.executeInstall(params)
// }

// func (s *service) executeDebugInstallAWSHello(params *params) (*model.CommandResponse, error) {
// 	manifest := aws_hello.Manifest()

// 	s.api.Mattermost.Log.Error(fmt.Sprintf("manifest = %v", manifest))
// 	app, _, err := s.api.Admin.ProvisionApp(
// 		&api.Context{
// 			ActingUserID: params.commandArgs.UserId,
// 		},
// 		api.SessionToken(params.commandArgs.Session.Token),
// 		&api.InProvisionApp{
// 			Manifest: manifest,
// 			Force:    true,
// 		},
// 	)
// 	s.api.Mattermost.Log.Error(fmt.Sprintf("app = %v", app))

// 	if err != nil {
// 		return errorOut(params, err)
// 	}

// 	conf := s.api.Configurator.GetConfig()

// 	// Finish the installation when the Dialog is submitted, see
// 	// <plugin>/http/dialog/install.go
// 	err = s.api.Mattermost.Frontend.OpenInteractiveDialog(
// 		dialog.NewInstallAppDialog(manifest, "", conf.PluginURL, params.commandArgs))
// 	if err != nil {
// 		return errorOut(params, errors.Wrap(err, "couldn't open an interactive dialog"))
// 	}

// 	s.api.Mattermost.Log.Error(fmt.Sprintf("before get team = %v", params.commandArgs.TeamId))

// 	team, err := s.api.Mattermost.Team.Get(params.commandArgs.TeamId)
// 	if err != nil {
// 		return errorOut(params, err)
// 	}
// 	s.api.Mattermost.Log.Error(fmt.Sprintf("after get team = %v", team))

// 	return &model.CommandResponse{
// 		GotoLocation: params.commandArgs.SiteURL + "/" + team.Name + "/messages/@" + app.BotUsername,
// 		Text:         fmt.Sprintf("%s. redirected to the DM with @%s to continue installing **%s**", "text", app.BotUsername, manifest.DisplayName),
// 		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
// 	}, nil
// }
