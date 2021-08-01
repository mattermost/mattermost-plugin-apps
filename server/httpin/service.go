package httpin

import (
	"net/http"

	"github.com/gorilla/mux"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/plugin"

	"github.com/mattermost/mattermost-plugin-apps/server/appservices"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type Service interface {
	ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request)
}

type service struct {
	router *mux.Router
}

var _ Service = (*service)(nil)

func NewService(router *mux.Router, mm *pluginapi.Client, log utils.Logger, conf config.Service, proxy proxy.Service, appServices appservices.Service,
	initf ...func(*mux.Router, *pluginapi.Client, utils.Logger, config.Service, proxy.Service, appservices.Service)) Service {
	for _, f := range initf {
		f(router, mm, log, conf, proxy, appServices)
	}
	router.Handle("{anything:.*}", http.NotFoundHandler())

	return &service{
		router: router,
	}
}

// Handle should be called by the plugin when a command invocation is received from the Mattermost server.
func (s *service) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	r.Header.Set("Mattermost-Session-Id", c.SessionId)
	r.Header.Set("Mattermost-Plugin-ID", c.SourcePluginId)
	s.router.ServeHTTP(w, r)
}
