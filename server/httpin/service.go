package httpin

import (
	"net/http"

	"github.com/gorilla/mux"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v6/plugin"

	"github.com/mattermost/mattermost-plugin-apps/server/appservices"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type Service interface {
	ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request)
}

type service struct {
	rh *Handler
}

var _ Service = (*service)(nil)

func NewService(mm *pluginapi.Client, config config.Service, log utils.Logger, session incoming.SessionService, router *mux.Router, proxy proxy.Service, appServices appservices.Service,
	initf ...func(*Handler, config.Service, proxy.Service, appservices.Service)) Service {
	rh := NewHandler(mm, config, log, session, router)

	for _, f := range initf {
		f(rh, config, proxy, appServices)
	}

	router.Handle("{anything:.*}", http.NotFoundHandler())

	return &service{
		rh: rh,
	}
}

// Handle should be called by the plugin when a command invocation is received from the Mattermost server.
func (s *service) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	r.Header.Set(config.MattermostSessionIDHeader, c.SessionId)
	s.rh.router.ServeHTTP(w, r)
}
