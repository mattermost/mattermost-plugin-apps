package dialog

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/server/appservices"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
	"github.com/mattermost/mattermost-plugin-apps/server/telemetry"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

const (
	InstallPath = "/install"
)

type dialog struct {
	mm    *pluginapi.Client
	log   utils.Logger
	proxy proxy.Service
	conf  config.Service
}

func Init(router *mux.Router, mm *pluginapi.Client, log utils.Logger, conf config.Service, proxy proxy.Service, _ appservices.Service, _ *telemetry.Telemetry) {
	d := dialog{
		mm:    mm,
		log:   log,
		proxy: proxy,
		conf:  conf,
	}

	subrouter := router.PathPrefix(config.InteractiveDialogPath).Subrouter()
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
