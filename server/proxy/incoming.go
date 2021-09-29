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
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

type Incoming struct {
	PluginID              string
	ActingUserID          string
	ActingUserAccessToken string
	AdminAccessToken      string
	SessionID             string
	SysAdminChecked       bool
}

func NewIncomingFromContext(cc apps.Context) Incoming {
	return Incoming{
		ActingUserID:          cc.ActingUserID,
		ActingUserAccessToken: cc.ActingUserAccessToken,
		AdminAccessToken:      cc.AdminAccessToken,
	}
}

func RequireUser(f func(http.ResponseWriter, *http.Request, Incoming)) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		actingUserID := req.Header.Get(config.MattermostUserIDHeader)
		sessionID := req.Header.Get(config.MattermostSessionIDHeader)
		if actingUserID == "" || sessionID == "" {
			httputils.WriteError(w, errors.Wrap(utils.ErrUnauthorized, "user ID and session ID are required"))
			return
		}

		f(w, req, Incoming{
			ActingUserID: actingUserID,
			SessionID:    sessionID,
		})
	}
}

func RequireSysadmin(mm *pluginapi.Client, f func(http.ResponseWriter, *http.Request, Incoming)) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		actingUserID := req.Header.Get(config.MattermostUserIDHeader)
		sessionID := req.Header.Get(config.MattermostSessionIDHeader)
		if actingUserID == "" || sessionID == "" {
			httputils.WriteError(w, errors.Wrap(utils.ErrUnauthorized, "user ID and session ID are required"))
			return
		}
		err := utils.EnsureSysAdmin(mm, actingUserID)
		if err != nil {
			httputils.WriteError(w, utils.ErrUnauthorized)
			return
		}

		f(w, req, Incoming{
			ActingUserID:    actingUserID,
			SessionID:       sessionID,
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
		ActingUserAccessToken: in.ActingUserAccessToken,
		AdminAccessToken:      in.AdminAccessToken,
	}
	return updated
}

func (in *Incoming) ensureUserToken(mm *pluginapi.Client) error {
	if in.ActingUserAccessToken != "" {
		return nil
	}
	if in.SessionID == "" {
		return utils.NewUnauthorizedError("no user token nor session ID")
	}
	session, err := utils.LoadSession(mm, in.SessionID, in.ActingUserID)
	if err != nil {
		return errors.Wrap(err, "failed to obtain user token from session")
	}
	in.ActingUserAccessToken = session.Token
	return nil
}

func (p *Proxy) getAdminClient(in Incoming) (mmclient.Client, error) {
	conf, mm, _ := p.conf.Basic()

	if in.AdminAccessToken == "" {
		if !in.SysAdminChecked {
			err := utils.EnsureSysAdmin(mm, in.ActingUserID)
			if err != nil {
				return nil, errors.Wrap(err, "failed to get admin access to Mattermost")
			}
		}
		err := in.ensureUserToken(mm)
		if err != nil {
			return nil, errors.Wrap(err, "failed to use the current user's token for admin access to Mattermost")
		}
		in.AdminAccessToken = in.ActingUserAccessToken
	}

	asAdmin := mmclient.NewHTTPClient(conf, in.AdminAccessToken)
	return asAdmin, nil
}
