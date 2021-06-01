// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/utils/md"
)

func (s *service) executeInfo(params *commandParams) (*model.CommandResponse, error) {
	conf := s.conf.GetConfig()
	resp := md.Markdownf("Mattermost Apps plugin version: %s, "+
		"[%s](https://github.com/mattermost/%s/commit/%s), built %s\n",
		conf.Version,
		conf.BuildHashShort,
		config.Repository,
		conf.BuildHash,
		conf.BuildDate)

	return out(params, resp)
}
