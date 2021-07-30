// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"net/http"

	pluginapi "github.com/mattermost/mattermost-plugin-api"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/mmclient"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

type Incoming struct {
	ActingUserID          string
	PluginID              string
	ActingUserAccessToken string
	AdminAccessToken      string
}

func RequireUser(f func(_ http.ResponseWriter, _ *http.Request, in Incoming)) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		actingUserID := req.Header.Get("Mattermost-User-Id")
		if actingUserID == "" {
			httputils.WriteError(w, utils.ErrUnauthorized)
			return
		}

		f(w, req, Incoming{
			ActingUserID: actingUserID,
		})
	}
}

func RequireUserToken(mm *pluginapi.Client, f func(_ http.ResponseWriter, _ *http.Request, in Incoming)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actingUserID := r.Header.Get("Mattermost-User-Id")
		sessionID := r.Header.Get("Mattermost-Session-ID")
		if actingUserID == "" || sessionID == "" {
			httputils.WriteError(w, utils.ErrUnauthorized)
			return
		}

		s, err := utils.LoadSession(mm, sessionID, actingUserID)
		if err != nil {
			httputils.WriteError(w, err)
		}
		f(w, r, Incoming{
			ActingUserID:          actingUserID,
			ActingUserAccessToken: s.Token,
		})
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
			sessionID := r.Header.Get("Mattermost-Session-ID")
			if actingUserID == "" || sessionID == "" {
				httputils.WriteError(w, utils.ErrUnauthorized)
				return
			}

			err := utils.EnsureSysAdmin(mm, actingUserID)
			if err != nil {
				httputils.WriteError(w, utils.ErrUnauthorized)
				return
			}
			s, err := utils.LoadSession(mm, sessionID, actingUserID)
			if err != nil {
				httputils.WriteError(w, err)
				return
			}
			in.ActingUserAccessToken = s.Token
			in.AdminAccessToken = s.Token
		}

		f(w, r, in)
	}
}

func RequireSysadmin(mm *pluginapi.Client, f func(_ http.ResponseWriter, _ *http.Request, in Incoming)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actingUserID := r.Header.Get("Mattermost-User-Id")
		in := Incoming{
			ActingUserID: actingUserID,
		}
		err := utils.EnsureSysAdmin(mm, actingUserID)
		if err != nil {
			httputils.WriteError(w, utils.ErrUnauthorized)
			return
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
	cc.ExpandedContext = apps.ExpandedContext{}
	cc.ActingUserID = in.ActingUserID
	cc.ActingUserAccessToken = in.ActingUserAccessToken
	cc.AdminAccessToken = in.AdminAccessToken
	return cc
}

func (in Incoming) newAppContext(app *apps.App, conf config.Config) apps.Context {
	cc := in.updateContext(apps.Context{})
	cc = forApp(app, cc, conf)
	return cc
}

// func (in Incoming) newContext() apps.Context {
// 	return in.updateContext(apps.Context{})
// }
