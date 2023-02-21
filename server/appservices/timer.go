// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package appservices

import (
	"context"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
)

type storedTimer struct {
	Call      apps.Call  `json:"call"`
	AppID     apps.AppID `json:"app_id"`
	UserID    string     `json:"user_id"`
	ChannelID string     `json:"channel_id,omitempty"`
	TeamID    string     `json:"team_id,omitempty"`
}

func (t storedTimer) Key(appID apps.AppID, at int64) string {
	return string(appID) + t.UserID + strconv.FormatInt(at, 10)
}

func (t storedTimer) Loggable() []interface{} {
	props := []interface{}{"user_id", t.UserID}
	props = append(props, "app_id", t.AppID)
	if t.ChannelID != "" {
		props = append(props, "call_team_id", t.TeamID)
	}
	if t.TeamID != "" {
		props = append(props, "call_channel_id", t.ChannelID)
	}
	return props
}

func (a *AppServices) CreateTimer(r *incoming.Request, t apps.Timer) error {
	err := r.Check(
		r.RequireActingUser,
		r.RequireSourceApp,
		t.Validate,
	)
	if err != nil {
		return err
	}

	st := storedTimer{
		Call:      t.Call,
		AppID:     r.SourceAppID(),
		UserID:    r.ActingUserID(),
		ChannelID: t.ChannelID,
		TeamID:    t.TeamID,
	}

	_, err = a.scheduler.ScheduleOnce(st.Key(r.SourceAppID(), t.At), time.UnixMilli(t.At), st)
	if err != nil {
		return errors.Wrap(err, "faild to schedule timer job")
	}

	return nil
}

func (a *AppServices) ExecuteTimer(key string, props interface{}) {
	t, ok := props.(storedTimer)
	if !ok {
		a.log.Debugw("Timer contained unknown props. Inoring the timer.", "key", key, "props", props)
		return
	}

	r := a.caller.NewIncomingRequest()

	r.Log = r.Log.With(t)

	ctx, cancel := context.WithTimeout(context.Background(), config.RequestTimeout)
	defer cancel()
	r = r.WithCtx(ctx)

	r = r.WithDestination(t.AppID)
	r = r.WithActingUserID(t.UserID)

	context := &apps.Context{
		UserAgentContext: apps.UserAgentContext{
			AppID:     t.AppID,
			TeamID:    t.TeamID,
			ChannelID: t.ChannelID,
		},
	}

	creq := apps.CallRequest{
		Call:    t.Call,
		Context: *context,
	}
	r.Log = r.Log.With(creq)
	_, cresp := a.caller.InvokeCall(r, creq)
	if cresp.Type == apps.CallResponseTypeError {
		if a.conf.Get().DeveloperMode {
			r.Log.WithError(cresp).Errorf("Timer execute failed")
		}
		return
	}
	r.Log = r.Log.With(cresp)

	r.Log.Debugf("Timer executed")
}
