// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"fmt"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/server/config"
)

func (s *service) executeInfo(params *commandParams) (*model.CommandResponse, error) {
	conf := s.conf.Get()
	resp := fmt.Sprintf("Mattermost Apps plugin version: %s, "+
		"[%s](https://github.com/mattermost/%s/commit/%s), built %s, Cloud Mode: %t, Developer Mode: %t\n",
		conf.Version,
		conf.BuildHashShort,
		config.Repository,
		conf.BuildHash,
		conf.BuildDate,
		conf.MattermostCloudMode,
		conf.DeveloperMode,
	)

	return out(params, resp)
}
