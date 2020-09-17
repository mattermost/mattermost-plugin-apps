// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"github.com/mattermost/mattermost-plugin-apps/server/constants"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
	"github.com/mattermost/mattermost-server/v5/model"
)

func (s *service) executeInfo(params *params) (*model.CommandResponse, error) {
	conf := s.apps.Config.GetConfig()
	resp := md.Markdownf("Mattermost Cloud Apps plugin version: %s, "+
		"[%s](https://github.com/mattermost/%s/commit/%s), built %s\n",
		conf.Version,
		conf.BuildHashShort,
		constants.Repository,
		conf.BuildHash,
		conf.BuildDate)

	return normalOut(params, resp, nil)
}
