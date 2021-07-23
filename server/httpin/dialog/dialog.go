package dialog

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-api/i18n"
	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/server/appservices"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
)

const (
	InstallPath = "/install"
)

type dialog struct {
	mm    *pluginapi.Client
	proxy proxy.Service
	conf  config.Service
	i18n  *i18n.Bundle
}

func Init(router *mux.Router, mm *pluginapi.Client, conf config.Service, proxy proxy.Service, _ appservices.Service, i18nBundle *i18n.Bundle) {
	d := dialog{
		mm:    mm,
		proxy: proxy,
		conf:  conf,
		i18n:  i18nBundle,
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
