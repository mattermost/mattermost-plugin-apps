// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"time"

	"github.com/hashicorp/go-multierror"

	"github.com/mattermost/mattermost-plugin-apps/utils"
)

// Timer s submitted by an app to the Timer API. It determines when
// the app would like to be notified, and how these notifications
// should be invoked.
type Timer struct {
	// At is the unix time in milliseconds when the timer should be executed.
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

	if time.Until(time.UnixMilli(t.At)) < 1*time.Second {
		result = multierror.Append(result, utils.NewInvalidError("at most be at least 1 second in the future"))
	}

	return result
}
