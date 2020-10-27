package helloapp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-api/experimental/oauther"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/constants"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

const AppSecret = "1234"

const (
	PathManifest                = "/mattermost-app.json"
	PathNotifyUserJoinedChannel = "/notify/" + string(api.SubjectUserJoinedChannel)
	PathInstall                 = "/form/install"
	PathConnectedInstall        = "/form/connected_install"
	PathPing                    = "/form/ping"
	PathSubmitEmbedded          = "/form/submit_embedded"
	PathCreateEmbedded          = "/form/create_embedded"
	PathOAuth2                  = "/oauth2"
	PathOAuth2Complete          = "/oauth2/complete" // /complete comes from OAuther
	PathLocations               = "/locations"
)

type helloapp struct {
	apps    *apps.Service
	OAuther oauther.OAuther
}

func Init(router *mux.Router, apps *apps.Service) {
	h := helloapp{
		apps: apps,
	}

	subrouter := router.PathPrefix(constants.HelloAppPath).Subrouter()

	subrouter.HandleFunc(PathManifest, h.handleManifest).Methods("GET")
	subrouter.PathPrefix(PathOAuth2).HandlerFunc(h.handleOAuth).Methods("GET")

	subrouter.HandleFunc(PathNotifyUserJoinedChannel, notify(h.handleUserJoinedChannel)).Methods("POST")

	subrouter.HandleFunc(PathInstall, call(h.handleInstall)).Methods("POST")
	subrouter.HandleFunc(PathConnectedInstall, call(h.handleConnectedInstall)).Methods("POST")
	subrouter.HandleFunc(PathPing, call(h.handlePing)).Methods("POST")
	subrouter.HandleFunc(PathSubmitEmbedded, call(h.handleSubmitEmbedded)).Methods("POST")
	subrouter.HandleFunc(PathCreateEmbedded, call(h.handleCreateEmbedded)).Methods("POST")

	subrouter.HandleFunc(PathLocations, CheckAuthentication(ExtractUserAndChannelID(h.HandleLocations))).Methods("GET")

	_ = h.InitOAuther()
}

func (h *helloapp) AppURL(path string) string {
	conf := h.apps.Configurator.GetConfig()
	return conf.PluginURL + constants.HelloAppPath + path
}

type CallHandler func(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, data *api.Call) (int, error)

func call(h CallHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		claims, err := checkJWT(req)
		if err != nil {
			httputils.WriteBadRequestError(w, err)
			return
		}

		data, err := api.UnmarshalCallFromReader(req.Body)
		if err != nil {
			httputils.WriteBadRequestError(w, err)
			return
		}

		statusCode, err := h(w, req, claims, data)
		if err != nil {
			httputils.WriteJSONError(w, statusCode, "", err)
			return
		}
	}
}

type notifyHandler func(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, data *api.Notification) (int, error)

func notify(h notifyHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		claims, err := checkJWT(req)
		if err != nil {
			httputils.WriteBadRequestError(w, err)
			return
		}

		data := api.Notification{}
		err = json.NewDecoder(req.Body).Decode(&data)
		if err != nil {
			httputils.WriteBadRequestError(w, err)
			return
		}

		statusCode, err := h(w, req, claims, &data)
		if err != nil {
			httputils.WriteJSONError(w, statusCode, "", err)
			return
		}
	}
}

func checkJWT(req *http.Request) (*apps.JWTClaims, error) {
	authValue := req.Header.Get(apps.OutgoingAuthHeader)
	if !strings.HasPrefix(authValue, "Bearer ") {
		return nil, errors.Errorf("missing %s: Bearer header", apps.OutgoingAuthHeader)
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
		return nil, err
	}

	return &claims, nil
}
