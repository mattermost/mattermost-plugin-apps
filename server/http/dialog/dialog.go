package dialog

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-server/v5/model"
)

const (
	InstallPath = "/install"
)

type dialog struct {
	apps *apps.Service
}

func Init(router *mux.Router, appsService *apps.Service) {
	d := dialog{
		apps: appsService,
	}

	subrouter := router.PathPrefix(apps.InteractiveDialogPath).Subrouter()
	subrouter.HandleFunc(InstallPath, d.handleInstall).Methods("POST")
}

func respondWithError(w http.ResponseWriter, status int, err error) {
	resp := model.SubmitDialogResponse{
		Error: errors.Wrap(err, "failed to install").Error(),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}
