// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"net/http"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/mmclient"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

type Incoming struct {
	ActingUserID          string
	PluginID              string
	ActingUserAccessToken string
	AdminAccessToken      string
	SessionID             string
}

func NewIncomingFromContext(cc apps.Context) Incoming {
	return Incoming{
		ActingUserID:          cc.ActingUserID,
		ActingUserAccessToken: cc.ActingUserAccessToken,
		AdminAccessToken:      cc.AdminAccessToken,
	}
}

func RequireUser(f func(_ http.ResponseWriter, _ *http.Request, in Incoming)) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		actingUserID := req.Header.Get("Mattermost-User-Id")
		sessionID := req.Header.Get("Mattermost-Session-Id")
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
		actingUserID := req.Header.Get("Mattermost-User-Id")
		sessionID := req.Header.Get("Mattermost-Session-Id")
		if actingUserID == "" || sessionID == "" {
			httputils.WriteError(w, errors.Wrap(utils.ErrUnauthorized, "user ID and session ID are required"))
			return
		}
		in := Incoming{
			ActingUserID: actingUserID,
			SessionID:    sessionID,
		}

		err := utils.EnsureSysAdmin(mm, actingUserID)
		if err != nil {
			httputils.WriteError(w, utils.ErrUnauthorized)
			return
		}

		f(w, req, in)
	}
}

func RequireSysadminOrPlugin(mm *pluginapi.Client, f func(_ http.ResponseWriter, _ *http.Request, in Incoming)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pluginID := r.Header.Get("Mattermost-Plugin-ID")
		actingUserID := r.Header.Get("Mattermost-User-Id")
		in := Incoming{
			PluginID:     pluginID,
			ActingUserID: actingUserID,
		}

		if pluginID == "" {
			sessionID := r.Header.Get("Mattermost-Session-Id")
			if actingUserID == "" || sessionID == "" {
				httputils.WriteError(w, errors.Wrap(utils.ErrUnauthorized, "user ID and session ID are required"))
				return
			}
			in.SessionID = sessionID

			err := utils.EnsureSysAdmin(mm, actingUserID)
			if err != nil {
				httputils.WriteError(w, utils.ErrUnauthorized)
				return
			}
		}

		f(w, r, in)
	}
}

func (p *Proxy) newSudoClient(in Incoming) mmclient.Client {
	conf := p.conf.GetConfig()
	var client mmclient.Client
	if in.PluginID != "" {
		client = mmclient.NewRPCClient(p.mm)
	} else {
		client = mmclient.NewHTTPClient(p.mm, conf, in.AdminAccessToken)
	}
	return client
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
