package restapi

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

type restapi struct {
	conf        config.Service
	proxy       proxy.Service
	appServices appservices.Service
}

func Init(router *mux.Router, conf config.Service, p proxy.Service, appServices appservices.Service, sessionService session.Service) {
	mm := conf.MattermostAPI()
	a := &restapi{
		conf:        conf,
		proxy:       p,
		appServices: appServices,
	}

	api := router.PathPrefix(path.API).Subrouter()

	c := request.NewContext(mm, conf, sessionService)

	a.initPing(api, c)

	// Proxy API, intended to be used by the user-agents (mobile, desktop, and
	// web).
	a.initCall(api, c)

	// User-agent APIs.
	a.initGetBindings(api, c)
	a.initGetBotIDs(api, c)
	a.initGetOAuthAppIDs(api, c)

	// App Service API, intended to be used by Apps. Subscriptions, KV, OAuth2
	// services.
	a.initSubscriptions(api, c)
	a.initKV(api, c)
	a.initOAuth2Store(api, c)

	// Admin API, can be used by plugins, external services, or the user agent.
	a.initAdmin(api, c)
	a.initGetApp(api, c)
	a.initMarketplace(api, c)
}

func appIDVar(r *http.Request) apps.AppID {
	s, ok := mux.Vars(r)["appid"]
	if ok {
		return apps.AppID(s)
	}
	return ""
}
