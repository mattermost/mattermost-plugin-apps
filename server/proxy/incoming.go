// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"net/http"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/pluginclient"
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
		pluginID := r.Header.Get(config.MattermostPluginIDHeader)
		actingUserID := r.Header.Get(config.MattermostUserIDHeader)
		in := Incoming{
			PluginID:     pluginID,
			ActingUserID: actingUserID,
		}

		if pluginID == "" {
			sessionID := r.Header.Get(config.MattermostSessionIDHeader)
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

func (p *Proxy) asAdmin(in Incoming) (Incoming, pluginclient.Client, error) {
	conf, mm, _ := p.conf.Basic()
	var client pluginclient.Client
	if in.PluginID != "" {
		client = pluginclient.NewRPCClient(mm)
	} else {
		if in.AdminAccessToken == "" && in.SessionID != "" {
			session, err := mm.Session.Get(in.SessionID)
			if err != nil {
				return in, nil, err
			}
			in.AdminAccessToken = session.Token
			in.ActingUserAccessToken = session.Token
		}

		client = pluginclient.NewHTTPClient(conf, in.AdminAccessToken)
	}
	return in, client, nil
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
