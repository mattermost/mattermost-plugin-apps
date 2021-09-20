package dialog

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/server/appservices"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
)

const (
	InstallPath = "/install"
)

type dialog struct {
	proxy proxy.Service
	conf  config.Service
}

func Init(router *mux.Router, conf config.Service, p proxy.Service, _ appservices.Service) {
	d := dialog{
		proxy: p,
		conf:  conf,
	}

	subrouter := router.PathPrefix(config.InteractiveDialogPath).Subrouter()

	subrouter.HandleFunc(InstallPath,
		proxy.RequireSysadminOrPlugin(conf.MattermostAPI(), d.handleInstall)).Methods("POST")
}

func respondWithError(w http.ResponseWriter, status int, err error) {
	resp := model.SubmitDialogResponse{
		Error: errors.Wrap(err, "failed to install").Error(),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}
