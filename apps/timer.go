// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"github.com/hashicorp/go-multierror"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

// Timer TOOD
type Timer struct {
	// At is the unix time in milliseconds when the timer should be executed
	At int64 `json:"at"`

	// Call is the (one-way) call to make upon the timers execution.
	Call Call `json:"call"`

	// ChannelID is a channel ID that is used for expansion of the Call (optional).
	ChannelID string `json:"channel_id,omitempty"`
	// TeamID is a team ID that is used for expansion of the Call (optional).
	TeamID string `json:"team_id,omitempty"`
}

func (t Timer) Validate() error {
	var result error
	emptyCall := Call{}
	if t.Call == emptyCall {
		result = multierror.Append(result, utils.NewInvalidError("call must not be empty"))
	}

	if t.At <= 0 {
		result = multierror.Append(result, utils.NewInvalidError("at must be positive"))
	}

	return result
}