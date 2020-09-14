package dialog

import (
	"github.com/gorilla/mux"

	apps "github.com/mattermost/mattermost-plugin-cloudapps/server/cloudapps"
	"github.com/mattermost/mattermost-plugin-cloudapps/server/constants"
)

const (
	InstallPath = "/install"
)

type dialog struct {
	apps *apps.Service
}

func Init(router *mux.Router, apps *apps.Service) {
	d := dialog{
		apps: apps,
	}

	subrouter := router.PathPrefix(constants.InteractiveDialogPath).Subrouter()
	subrouter.HandleFunc("/install", d.handleInstall).Methods("POST")
}
