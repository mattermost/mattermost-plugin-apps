package httpin

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-server/v6/plugin"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/appservices"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

const AppIDVar = "appid"
const AppIDPath = "/{appid}"

type Service struct {
	AppServices appservices.Service
	Config      config.Service
	Proxy       proxy.Service

	baseLog     utils.Logger
	router      *mux.Router
	handlerFunc handlerFunc
}

var _ http.Handler = (*Service)(nil)

type handlerFunc func(*incoming.Request, http.ResponseWriter, *http.Request)

func NewService(proxy proxy.Service, appservices appservices.Service, conf config.Service, log utils.Logger) *Service {
	rootHandler := &Service{
		AppServices: appservices,
		Config:      conf,
		Proxy:       proxy,
		baseLog:     log,
		router:      mux.NewRouter(),
	}
	rootHandler.router.Handle("{anything:.*}", http.NotFoundHandler())

	// Set up the "gateway" endpoints (APIs and pages/files) in the app's namespace, /{appid}/...
	h := rootHandler.PathPrefix(AppIDPath)

	// Static files.
	h.HandleFunc(path.Static+"/{name}", h.Static).Methods(http.MethodGet)

	// Incoming remote webhooks.
	h.HandleFunc(path.Webhook, h.Webhook).Methods(http.MethodPost)
	h.HandleFunc(path.Webhook+"/{path}", h.Webhook).Methods(http.MethodPost)

	// Remote OAuth2: /{appid}/oauth2/remote/connect and /{appid}/oauth2/remote/complete
	h.HandleFunc(path.RemoteOAuth2Connect, h.RemoteOAuth2Connect).Methods(http.MethodGet)
	h.HandleFunc(path.RemoteOAuth2Complete, h.RemoteOAuth2Complete).Methods(http.MethodGet)

	// Set up the REST APIs
	h = rootHandler.PathPrefix(path.API)

	// Ping.
	h.HandleFunc(path.Ping, h.Ping).Methods(http.MethodPost)

	// User-agent APIs.
	h.HandleFunc(path.Call, h.Call).Methods(http.MethodPost)
	h.HandleFunc(path.Bindings, h.GetBindings).Methods(http.MethodGet)
	h.HandleFunc(path.BotIDs, h.GetBotIDs).Methods(http.MethodGet)
	h.HandleFunc(path.OAuthAppIDs, h.GetOAuthAppIDs).Methods(http.MethodGet)

	// App Service API, intended to be used by Apps. Subscriptions, KV, OAuth2
	// services.
	h.HandleFunc(path.KV+"/{key}", h.KVDelete).Methods(http.MethodDelete)
	h.HandleFunc(path.KV+"/{key}", h.KVGet).Methods(http.MethodGet)
	h.HandleFunc(path.KV+"/{key}", h.KVPut).Methods(http.MethodPut, http.MethodPost)
	h.HandleFunc(path.KV+"/{prefix}/{key}", h.KVDelete).Methods(http.MethodDelete)
	h.HandleFunc(path.KV+"/{prefix}/{key}", h.KVGet).Methods(http.MethodGet)
	h.HandleFunc(path.KV+"/{prefix}/{key}", h.KVPut).Methods(http.MethodPut, http.MethodPost)
	h.HandleFunc(path.OAuth2App, h.OAuth2StoreApp).Methods(http.MethodPut, http.MethodPost)
	h.HandleFunc(path.OAuth2User, h.OAuth2GetUser).Methods(http.MethodGet)
	h.HandleFunc(path.OAuth2User, h.OAuth2StoreUser).Methods(http.MethodPut, http.MethodPost)
	h.HandleFunc(path.Subscribe, h.GetSubscriptions).Methods(http.MethodGet)
	h.HandleFunc(path.Subscribe, h.Subscribe).Methods(http.MethodPost)
	h.HandleFunc(path.Unsubscribe, h.Unsubscribe).Methods(http.MethodPost)

	// Admin API, can be used by plugins, external services, or the user agent.
	h.HandleFunc(path.DisableApp, h.DisableApp).Methods(http.MethodPost)
	h.HandleFunc(path.EnableApp, h.EnableApp).Methods(http.MethodPost)
	h.HandleFunc(path.InstallApp, h.InstallApp).Methods(http.MethodPost)
	h.HandleFunc(path.Marketplace, h.GetMarketplace).Methods(http.MethodGet)
	h.HandleFunc(path.UninstallApp, h.UninstallApp).Methods(http.MethodPost)
	h.HandleFunc(path.UpdateAppListing, h.UpdateAppListing).Methods(http.MethodPost)
	h.PathPrefix(path.Apps).PathPrefix(`/{appid:[A-Za-z0-9-_.]+}`).HandleFunc("", h.GetApp).Methods(http.MethodGet)

	return rootHandler
}

// clone creates a shallow copy of Handler, allowing clones to apply changes per
// handler func. Don't copy the following fields as they are specific to the
// handler func: handlerFunc, checks.
func (s *Service) clone() *Service {
	clone := *s
	clone.handlerFunc = nil
	return &clone
}

func (s *Service) PathPrefix(prefix string) *Service {
	clone := s.clone()
	clone.router = clone.router.PathPrefix(prefix).Subrouter()
	return clone
}

func (s *Service) HandleFunc(path string, handlerFunc handlerFunc) *mux.Route {
	clone := s.clone()
	clone.handlerFunc = handlerFunc
	return clone.router.Handle(path, clone)
}

// ServePluginHTTP is the interface invoked from the plugin's ServeHTTP
func (s *Service) ServePluginHTTP(c *plugin.Context, w http.ResponseWriter, req *http.Request) {
	req.Header.Set(config.MattermostSessionIDHeader, c.SessionId)
	s.router.ServeHTTP(w, req)
}

// ServeHTTP is the go http.Handler (mux compliant).
func (s *Service) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Setup the incoming request.
	r := s.Proxy.NewIncomingRequest()
	r = r.WithActingUserID(req.Header.Get(config.MattermostUserIDHeader))
	r = r.WithSourcePluginID(req.Header.Get(config.MattermostPluginIDHeader))
	r = r.WithSessionID(req.Header.Get(config.MattermostSessionIDHeader))
	if s, ok := mux.Vars(req)[AppIDVar]; ok {
		r = r.WithDestination(apps.AppID(s))
	}
	var cancel context.CancelFunc
	r = r.WithTimeout(config.RequestTimeout, &cancel)
	defer cancel()
	r.Log = r.Log.With(
		"path", req.URL.Path,
	)

	// Output panics in dev. mode.
	defer func() {
		if x := recover(); x != nil {
			stack := string(debug.Stack())

			r.Log.Errorw(
				"Recovered from a panic in an HTTP handler",
				"url", req.URL.String(),
				"error", x,
				"stack", string(debug.Stack()),
			)

			txt := "Panicked while handling the request. "

			if s.Config.Get().DeveloperMode {
				txt += fmt.Sprintf("Error: %v. Stack: %v", x, stack)
			} else {
				txt += "Please check the server logs for more details."
			}

			http.Error(w, txt, http.StatusInternalServerError)
		}
	}()

	s.handlerFunc(r, w, req)

	if s.Config.Get().DeveloperMode {
		r.Log.Debugf("HTTP")
	}
}
