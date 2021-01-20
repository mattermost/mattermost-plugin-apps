package restapi

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

type SubscribeResponse struct {
	Error  string            `json:"error,omitempty"`
	Errors map[string]string `json:"errors,omitempty"`
}

type restapi struct {
	api *api.Service
}

func Init(router *mux.Router, appsService *api.Service) {
	a := &restapi{
		api: appsService,
	}

	subrouter := router.PathPrefix(api.APIPath).Subrouter()
	subrouter.HandleFunc(api.BindingsPath, checkAuthorized(a.handleGetBindings)).Methods("GET")
	subrouter.HandleFunc(api.CallPath, a.handleCall).Methods("POST")
	subrouter.HandleFunc(api.SubscribePath, a.handleSubscribe).Methods("POST", "DELETE")

	subrouter.HandleFunc(api.KVPath+"/{key}", a.handleKV(a.kvGet)).Methods("GET")
	subrouter.HandleFunc(api.KVPath+"/{key}", a.handleKV(a.kvPut)).Methods("PUT", "POST")
	subrouter.HandleFunc(api.KVPath+"/", a.handleKV(a.kvList)).Methods("GET")
	subrouter.HandleFunc(api.KVPath+"/{key}", a.handleKV(a.kvHead)).Methods("HEAD")
	subrouter.HandleFunc(api.KVPath+"/{key}", a.handleKV(a.kvDelete)).Methods("DELETE")

	subrouter.HandleFunc(api.AssetPath, a.assetGet).Methods("GET")
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
