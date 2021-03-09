package dialog

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
)

const (
	InstallPath = "/install"
)

type dialog struct {
	mm    *pluginapi.Client
	admin api.Admin
}

func Init(router *mux.Router, mm *pluginapi.Client, _ api.Configurator, _ api.Proxy, admin api.Admin, _ api.AppServices) {
	d := dialog{
		mm:    mm,
		admin: admin,
	}

	subrouter := router.PathPrefix(api.InteractiveDialogPath).Subrouter()
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
