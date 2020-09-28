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

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/constants"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

const AppSecret = "1234"

const (
	PathManifest             = "/mattermost-app.json"
	PathWishInstall          = "/wish/install"
	PathWishConnectedInstall = "/wish/connected_install"
	PathOAuth2               = "/oauth2"
	PathOAuth2Complete       = "/oauth2/complete" // /complete comes from OAuther
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

	subrouter.HandleFunc(PathWishInstall, wish(h.handleInstall)).Methods("POST")
	subrouter.HandleFunc(PathWishConnectedInstall, wish(h.handleConnectedInstall)).Methods("POST")

	subrouter.PathPrefix(PathOAuth2).HandlerFunc(h.handleOAuth).Methods("POST")

	_ = h.InitOAuther()
}

func (h *helloapp) AppURL(path string) string {
	conf := h.apps.Configurator.GetConfig()
	return conf.PluginURL + constants.HelloAppPath + path
}

type WishHandler func(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, data *apps.CallData) (int, error)

func wish(wishHandler WishHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
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

		statusCode, err := wishHandler(w, req, &claims, &data)
		if err != nil {
			httputils.WriteJSONError(w, statusCode, "", err)
			return
		}
	}
}
