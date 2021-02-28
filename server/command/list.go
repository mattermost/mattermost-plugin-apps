// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"github.com/mattermost/mattermost-server/v5/model"
)

func (s *service) executeList(params *params) (*model.CommandResponse, error) {
	_, txt, err := s.admin.ListApps()
	if err != nil {
		return errorOut(params, err)
	}
	return out(params, txt)
}
