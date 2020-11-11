package restapi

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

type SubscribeResponse struct {
	Error  string            `json:"error,omitempty"`
	Errors map[string]string `json:"errors,omitempty"`
}

type restapi struct {
	mm   *pluginapi.Client
	apps *apps.Service
}

func Init(router *mux.Router, appsService *apps.Service) {
	a := &restapi{
		mm:   appsService.Mattermost,
		apps: appsService,
	}

	subrouter := router.PathPrefix(apps.APIPath).Subrouter()
	subrouter.HandleFunc(apps.BindingsPath, checkAuthorized(a.handleGetBindings)).Methods("GET")
	subrouter.HandleFunc(apps.CallPath, a.handleCall).Methods("POST")
	subrouter.HandleFunc(apps.SubscribePath, a.handleSubscribe).Methods("POST", "DELETE")
}

func checkAuthorized(f func(http.ResponseWriter, *http.Request, string)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		actingUserID := req.Header.Get("Mattermost-User-Id")
		if actingUserID == "" {
			httputils.WriteUnauthorizedError(w, errors.New("not authorized"))
			return
		}

		f(w, req, actingUserID)
	}
}
