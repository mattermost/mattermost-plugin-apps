package proxy

import (
	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/constants"
)

const (
	CallPath = "/call"
)

type proxy struct {
	apps *apps.Service
}

func Init(router *mux.Router, apps *apps.Service) {
	w := proxy{
		apps: apps,
	}

	subrouter := router.PathPrefix(constants.ProxyPath).Subrouter()
	subrouter.HandleFunc(CallPath, w.handleCall).Methods("POST")
}
