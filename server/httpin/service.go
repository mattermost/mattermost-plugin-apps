package httpin

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-server/v6/plugin"

	"github.com/mattermost/mattermost-plugin-apps/server/appservices"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
)

type Service interface {
	ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request)
}

type service struct {
	router *mux.Router
}

var _ Service = (*service)(nil)

func NewService(router *mux.Router, conf config.Service, proxy proxy.Service, appServices appservices.Service,
	initf ...func(*mux.Router, config.Service, proxy.Service, appservices.Service)) Service {
	for _, f := range initf {
		f(router, conf, proxy, appServices)
	}
	router.Handle("{anything:.*}", http.NotFoundHandler())

	return &service{
		router: router,
	}
}

// Handle should be called by the plugin when a command invocation is received from the Mattermost server.
func (s *service) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	r.Header.Set("MM_SESSION_ID", c.SessionId)
	s.router.ServeHTTP(w, r)
}
