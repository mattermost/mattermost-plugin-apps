// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"net/http"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/mmclient"
	"github.com/mattermost/mattermost-plugin-apps/server/session"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

type Incoming struct {
	PluginID              string
	ActingUserID          string
	actingUserAccessToken string
	SysAdminChecked       bool
	sessionService        session.Service
}

func NewIncomingFromContext(cc apps.Context) Incoming {
	return Incoming{
		ActingUserID:          cc.ActingUserID,
		actingUserAccessToken: cc.ActingUserAccessToken,
	}
}

func RequireUser(f func(http.ResponseWriter, *http.Request, Incoming)) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		actingUserID := req.Header.Get(config.MattermostUserIDHeader)
		if actingUserID == "" {
			httputils.WriteError(w, errors.Wrap(utils.ErrUnauthorized, "user ID is required"))
			return
		}

		f(w, req, Incoming{
			ActingUserID: actingUserID,
		})
	}
}

func RequireSysadmin(mm *pluginapi.Client, f func(http.ResponseWriter, *http.Request, Incoming)) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		actingUserID := req.Header.Get(config.MattermostUserIDHeader)
		if actingUserID == "" {
			httputils.WriteError(w, errors.Wrap(utils.ErrUnauthorized, "user ID is required"))
			return
		}
		err := utils.EnsureSysAdmin(mm, actingUserID)
		if err != nil {
			httputils.WriteError(w, utils.ErrUnauthorized)
			return
		}

		f(w, req, Incoming{
			ActingUserID:    actingUserID,
			SysAdminChecked: true,
		})
	}
}

func RequireSysadminOrPlugin(mm *pluginapi.Client, f func(http.ResponseWriter, *http.Request, Incoming)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pluginID := r.Header.Get(config.MattermostPluginIDHeader)
		if pluginID != "" {
			f(w, r, Incoming{
				PluginID: pluginID,
			})
			return
		}

		RequireSysadmin(mm, f)(w, r)
	}
}

func (in Incoming) updateContext(cc apps.Context) apps.Context {
	updated := cc
	updated.ActingUserID = in.ActingUserID
	updated.ExpandedContext = apps.ExpandedContext{
		ActingUserAccessToken: in.actingUserAccessToken,
	}
	return updated
}

func (in Incoming) UserAccessToken() (string, error) {
	if in.actingUserAccessToken != "" {
		return in.actingUserAccessToken, nil
	}

	// TODO

	return "", nil
}

func (p *Proxy) getClient(in Incoming) (mmclient.Client, error) {
	conf, _, _ := p.conf.Basic()

	token, err := in.UserAccessToken()
	if err != nil {
		return nil, errors.Wrap(err, "failed to use the current user's token for admin access to Mattermost")
	}

	return mmclient.NewHTTPClient(conf, token), nil
}
