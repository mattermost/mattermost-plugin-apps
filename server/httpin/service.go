package httpin

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-server/v6/plugin"

	"github.com/mattermost/mattermost-plugin-apps/server/appservices"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy/request"
)

type Service interface {
	ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request)
}

type service struct {
	router *mux.Router
}

var _ Service = (*service)(nil)

func NewService(c *request.Context, router *mux.Router, proxy proxy.Service, appServices appservices.Service,
	initf ...func(*request.Context, *mux.Router, proxy.Service, appservices.Service)) Service {
	for _, f := range initf {
		f(c, router, proxy, appServices)
	}
	router.Use(recoveryHandler(c.Clone()))
	router.Handle("{anything:.*}", http.NotFoundHandler())

	return &service{
		router: router,
	}
}
func recoveryHandler(c *request.Context) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if x := recover(); x != nil {
					stack := string(debug.Stack())

					c.Log.Errorw(
						"Recovered from a panic in an HTTP handler",
						"url", r.URL.String(),
						"error", x,
						"stack", string(debug.Stack()),
					)

					txt := "Paniced while handling the request. "

					if c.Config().Get().DeveloperMode {
						txt += fmt.Sprintf("Error: %v. Stack: %v", x, stack)
					} else {
						txt += "Please check the server logs for more details."
					}

					http.Error(w, txt, http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// Handle should be called by the plugin when a command invocation is received from the Mattermost server.
func (s *service) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
