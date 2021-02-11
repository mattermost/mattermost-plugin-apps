package restapi

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

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
	subrouter.HandleFunc(api.SubscribePath, a.handleSubscribe).Methods("POST")
	subrouter.HandleFunc(api.UnsubscribePath, a.handleUnsubscribe).Methods("POST")

	subrouter.HandleFunc(api.KVPath+"/{key}", a.handleKV(a.kvGet)).Methods("GET")
	subrouter.HandleFunc(api.KVPath+"/{key}", a.handleKV(a.kvPut)).Methods("PUT", "POST")
	subrouter.HandleFunc(api.KVPath+"/", a.handleKV(a.kvList)).Methods("GET")
	subrouter.HandleFunc(api.KVPath+"/{key}", a.handleKV(a.kvHead)).Methods("HEAD")
	subrouter.HandleFunc(api.KVPath+"/{key}", a.handleKV(a.kvDelete)).Methods("DELETE")

	subrouter.HandleFunc(api.PathMarketplace, checkAuthorized(a.handleGetMarketplace)).Methods(http.MethodGet)
}

func checkAuthorized(f func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		actingUserID := req.Header.Get("Mattermost-User-Id")
		if actingUserID == "" {
			httputils.WriteUnauthorizedError(w, errors.New("not authorized"))
			return
		}

		f(w, req, actingUserID)
	}
}
