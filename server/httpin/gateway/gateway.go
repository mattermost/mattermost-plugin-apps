package gateway

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/appservices"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy/request"
	"github.com/mattermost/mattermost-plugin-apps/server/session"
)

type gateway struct {
	conf  config.Service
	proxy proxy.Service
}

func Init(router *mux.Router, conf config.Service, p proxy.Service, _ appservices.Service, sessionService session.Service) {
	mm := conf.MattermostAPI()
	g := &gateway{
		conf:  conf,
		proxy: p,
	}

	subrouter := router.PathPrefix(path.Apps).Subrouter()

	c := request.NewContext(mm, conf, sessionService)

	// Static
	subrouter.Handle("/{appid}/"+path.StaticFolder+"/{name}",
		request.AddContext(g.static, c).RequireUser()).Methods(http.MethodGet)

	// Incoming remote webhooks
	subrouter.Handle("/{appid}"+path.Webhook,
		request.AddContext(g.handleWebhook, c)).Methods(http.MethodPost)
	subrouter.Handle("/{appid}"+path.Webhook+"/{path}",
		request.AddContext(g.handleWebhook, c)).Methods(http.MethodPost)

	// Remote OAuth2
	subrouter.Handle("/{appid}"+path.RemoteOAuth2Connect,
		request.AddContext(g.remoteOAuth2Connect, c).RequireUser()).Methods(http.MethodGet)
	subrouter.Handle("/{appid}"+path.RemoteOAuth2Complete,
		request.AddContext(g.remoteOAuth2Complete, c).RequireUser()).Methods(http.MethodGet)
}

func appIDVar(r *http.Request) apps.AppID {
	s, ok := mux.Vars(r)["appid"]
	if ok {
		return apps.AppID(s)
	}
	return ""
}
