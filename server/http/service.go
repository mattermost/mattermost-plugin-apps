package http

import (
	"net/http"

	"github.com/gorilla/mux"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/plugin"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
)

type Service interface {
	ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request)
}

type service struct {
	router *mux.Router
}

var _ Service = (*service)(nil)

func NewService(router *mux.Router, mm *pluginapi.Client, conf api.Configurator, proxy api.Proxy, admin api.Admin, appServices api.AppServices,
	initf ...func(*mux.Router, *pluginapi.Client, api.Configurator, api.Proxy, api.Admin, api.AppServices)) Service {
	for _, f := range initf {
		f(router, mm, conf, proxy, admin, appServices)
	}
	router.Handle("{anything:.*}", http.NotFoundHandler())

	return &service{
		router: router,
	}
}

// Handle should be called by the plugin when a command invocation is received from the Mattermost server.
func (s *service) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	r.Header.Add("MM_SESSION_ID", c.SessionId)
	s.router.ServeHTTP(w, r)
}
