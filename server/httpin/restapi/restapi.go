package restapi

import (
	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/appservices"
	"github.com/mattermost/mattermost-plugin-apps/server/httpin/handler"
)

type restapi struct {
	*handler.Handler
	appServices appservices.Service
}

func Init(h *handler.Handler, appServices appservices.Service) {
	a := &restapi{
		Handler:     h,
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
