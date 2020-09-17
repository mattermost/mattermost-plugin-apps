// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"github.com/mattermost/mattermost-plugin-apps/server/constants"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

func (c *command) handleInfo(parameters []string) (md.MD, error) {
	conf := c.Configurator.Get()
	resp := md.Markdownf("Mattermost Cloud Apps plugin version: %s, "+
		"[%s](https://github.com/mattermost/%s/commit/%s), built %s\n",
		conf.Version,
		conf.BuildHashShort,
		constants.Repository,
		conf.BuildHash,
		conf.BuildDate)

	return resp, nil
}
