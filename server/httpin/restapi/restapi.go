package restapi

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/appservices"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/httpin"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
)

type restapi struct {
	conf        config.Service
	proxy       proxy.Service
	appServices appservices.Service
}

func Init(h *httpin.Handler, conf config.Service, p proxy.Service, appServices appservices.Service) {
	a := &restapi{
		conf:        conf,
		proxy:       p,
		appServices: appServices,
	}

	h = h.PathPrefix(path.API)

	a.initPing(h)

	// Proxy API, intended to be used by the user-agents (mobile, desktop, and
	// web).
	a.initCall(h)

	// User-agent APIs.
	a.initGetBindings(h)
	a.initGetBotIDs(h)
	a.initGetOAuthAppIDs(h)

	// App Service API, intended to be used by Apps. Subscriptions, KV, OAuth2
	// services.
	a.initSubscriptions(h)
	a.initKV(h)
	a.initOAuth2Store(h)

	// Admin API, can be used by plugins, external services, or the user agent.
	a.initAdmin(h)
	a.initGetApp(h)
	a.initMarketplace(h)
}

func appIDVar(r *http.Request) apps.AppID {
	s, ok := mux.Vars(r)["appid"]
	if ok {
		return apps.AppID(s)
	}
	return ""
}
