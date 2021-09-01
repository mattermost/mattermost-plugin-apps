package httpin

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gorilla/mux"

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

func NewService(router *mux.Router, conf config.Service, proxy proxy.Service, appServices appservices.Service,
	initf ...func(*mux.Router, config.Service, proxy.Service, appservices.Service)) Service {
	for _, f := range initf {
		f(router, conf, proxy, appServices)
	}
	router.Use(recoveryHandler(conf.Logger(), conf.Get().DeveloperMode))
	router.Handle("{anything:.*}", http.NotFoundHandler())

	return &service{
		router: router,
	}
}
func recoveryHandler(log utils.Logger, developerMode bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func(log utils.Logger, developerMode bool) {
				if x := recover(); x != nil {
					stack := string(debug.Stack())

					log.Errorw(
						"Recovered from a panic in an HTTP handler",
						"url", r.URL.String(),
						"error", x,
						"stack", string(debug.Stack()),
					)

					txt := "Paniced while handling the request. "

					if developerMode {
						txt += fmt.Sprintf("Error: %v. Stack: %v", x, stack)
					} else {
						txt += "Please check the server logs for more details."
					}

					http.Error(w, txt, http.StatusInternalServerError)
				}
			}(log, developerMode)

			next.ServeHTTP(w, r)
		})
	}
}

// Handle should be called by the plugin when a command invocation is received from the Mattermost server.
func (s *service) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	r.Header.Set(config.MattermostSessionIDHeader, c.SessionId)
	s.router.ServeHTTP(w, r)
}
