// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"net/http"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v6/model"
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

func RequireSysadmin(mm *pluginapi.Client, f func(_ http.ResponseWriter, _ *http.Request, in Incoming)) http.HandlerFunc {
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

		in := Incoming{
			ActingUserID:    actingUserID,
			SessionID:       sessionID,
			SysAdminChecked: true,
		}
		f(w, req, in)
	}
}

func RequireSysadminOrPlugin(mm *pluginapi.Client, f func(_ http.ResponseWriter, _ *http.Request, in Incoming)) http.HandlerFunc {
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

func (in *Incoming) ensureUserTokens(mm *pluginapi.Client, adminRequested bool) error {
	var session *model.Session
	var err error
	if in.ActingUserAccessToken == "" && in.SessionID != "" {
		session, err = utils.LoadSession(mm, in.SessionID, in.ActingUserID)
		if err != nil {
			return err
		}
		in.ActingUserAccessToken = session.Token
	}
	if in.ActingUserAccessToken == "" {
		return errors.New("failed to obtain the acting user token")
	}

	if adminRequested {
		if !in.SysAdminChecked {
			err = utils.EnsureSysAdmin(mm, in.ActingUserID)
			if err != nil {
				return err
			}
		}
		in.AdminAccessToken = in.ActingUserAccessToken
	}
	return err
}

func (in Incoming) getExpandClient(app apps.App, conf config.Config, mm *pluginapi.Client) (mmclient.Client, error) {
	switch {
	case app.GrantedPermissions.Contains(apps.PermissionActAsAdmin):
		// If the app has admin permission anyway, use the RPC client for performance reasons
		return mmclient.NewRPCClient(mm), nil

	case app.GrantedPermissions.Contains(apps.PermissionActAsUser) && in.ActingUserID != "":
		// The OAuth2 token should be used here once it's implemented
		err := in.ensureUserTokens(mm, true)
		if err != nil {
			return nil, err
		}
		return mmclient.NewHTTPClient(conf, in.ActingUserAccessToken), nil

	case app.GrantedPermissions.Contains(apps.PermissionActAsBot):
		return mmclient.NewHTTPClient(conf, app.BotAccessToken), nil

	default:
		return nil, utils.NewUnauthorizedError("apps without any ActAs* permission can't expand")
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
