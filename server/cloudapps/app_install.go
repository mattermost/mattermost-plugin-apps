// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package cloudapps

import (
	"errors"

	"github.com/mattermost/mattermost-plugin-cloudapps/server/utils/md"
)

type InInstallApp struct {
	App App
}

type OutInstallApp struct {
	md.MD
}

func (r *registry) InstallApp(in *InInstallApp) (*OutInstallApp, error) {
	if in.App.AppID == "" {
		return nil, errors.New("app ID must not be empty")
	}
	r.apps[in.App.AppID] = &in.App

	out := &OutInstallApp{
		MD: md.Markdownf("Installed Cloud App %s (%s)", in.App.DisplayName, in.App.AppID),
	}
	return out, nil
}
