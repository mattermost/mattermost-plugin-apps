package helloapp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/configurator"
	"github.com/mattermost/mattermost-plugin-apps/server/constants"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

const AppSecret = "1234"

type helloapp struct {
	mm           *pluginapi.Client
	configurator configurator.Service
}

func Init(router *mux.Router, apps *apps.Service) {
	a := helloapp{
		mm:           apps.Mattermost,
		configurator: apps.Configurator,
	}

	subrouter := router.PathPrefix(constants.HelloAppPath).Subrouter()
	subrouter.HandleFunc("/mattermost-app.json", a.handleManifest).Methods("GET")

	subrouter.HandleFunc("/wish/install", a.handleInstall).Methods("POST")
}

func (h *helloapp) handleManifest(w http.ResponseWriter, req *http.Request) {
	conf := h.configurator.GetConfig()

	rootURL := conf.PluginURL + constants.HelloAppPath

	httputils.WriteJSON(w,
		apps.Manifest{
			AppID:       "hello",
			DisplayName: "Hallo სამყარო",
			Description: "Hallo სამყარო test app",
			RootURL:     rootURL,
			RequestedPermissions: []apps.PermissionType{
				apps.PermissionUserJoinedChannelNotification,
				apps.PermissionActAsUser,
				apps.PermissionActAsBot,
			},
			Install: &apps.Wish{
				URL: rootURL + "/wish/install",
			},
			CallbackURL: rootURL + "/oauth",
			Homepage:    rootURL,
		})
}

func (h *helloapp) handleInstall(w http.ResponseWriter, req *http.Request) {
	authValue := req.Header.Get(apps.OutgoingAuthHeader)
	if !strings.HasPrefix(authValue, "Bearer ") {
		httputils.WriteBadRequestError(w, errors.Errorf("missing %s: Bearer header", apps.OutgoingAuthHeader))
		return
	}

	jwtoken := strings.TrimPrefix(authValue, "Bearer ")
	claims := apps.JWTClaims{}
	_, err := jwt.ParseWithClaims(jwtoken, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(AppSecret), nil
	})
	if err != nil {
		httputils.WriteBadRequestError(w, err)
		return
	}

	data := apps.CallData{}
	err = json.NewDecoder(req.Body).Decode(&data)
	if err != nil {
		httputils.WriteBadRequestError(w, err)
		return
	}

	// The freshly created bot token is largely useless, so we need the acting
	// user (sysadmin) to OAuth2 connect first. This can be done after OAuth2
	// (OAuther) is fully integrated.

	// TODO Install: create channel, subscribe, etc.

	httputils.WriteJSON(w,
		apps.CallResponse{
			Type:     apps.ResponseTypeOK,
			Markdown: "Installed! <><>",
			Data:     map[string]interface{}{"status": "ok"},
		})
}
