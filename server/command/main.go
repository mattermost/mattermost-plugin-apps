// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

func (c *command) handleMain(parameters []string) (md.MD, error) {
	subcommands := map[string]func([]string) (md.MD, error){
		"info":        c.handleInfo,
		"debug-clean": c.handleDebugClean,
	}

	return runSubcommand(subcommands, parameters)
}

func (c *command) handleDebugClean(parameters []string) (md.MD, error) {
	return "<><> TODO", nil
}
