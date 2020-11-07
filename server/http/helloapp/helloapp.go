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

const (
	AppID          = "hello"
	AppSecret      = "1234"
	AppDisplayName = "Hallo სამყარო"
	AppDescription = "Hallo სამყარო test app"
)

const (
	fieldUserID   = "userID"
	fieldMessage  = "message"
	fieldResponse = "response"
)

const (
	PathManifest       = "/mattermost-app.json"
	PathInstall        = constants.AppInstallPath  // convention for Mattermost Apps
	PathBindings       = constants.AppBindingsPath // convention for Mattermost Apps
	PathOAuth2         = "/oauth2"                 // convention for Mattermost Apps, comes from OAuther
	PathOAuth2Complete = "/oauth2/complete"        // convention for Mattermost Apps, comes from OAuther

	PathConnectedInstall = "/connected_install"
	PathSendSurvey       = "/send"
	PathSubscribeChannel = "/subscribe"
	PathSurvey           = "/survey"

	PathNotifyUserJoinedChannel = "/notify-user-joined-channel"
)

type helloapp struct {
	apps    *apps.Service
	OAuther oauther.OAuther
}

// Init hello app router
func Init(router *mux.Router, apps *apps.Service) {
	h := helloapp{
		apps: apps,
	}

	r := router.PathPrefix(constants.HelloAppPath).Subrouter()
	r.HandleFunc(PathManifest, h.handleManifest).Methods("GET")
	r.PathPrefix(PathOAuth2).HandlerFunc(h.handleOAuth).Methods("GET")
	handleGetWithContext(r, PathBindings, h.handleBindings)

	handleCall(r, PathInstall, h.fInstall)
	handleCall(r, PathConnectedInstall, h.fConnectedInstall)
	handleCall(r, PathSendSurvey, h.fSendSurvey)
	handleCall(r, PathSurvey, h.fSurvey)

	handleNotify(r, PathNotifyUserJoinedChannel, h.handleUserJoinedChannel)

	_ = h.InitOAuther()
}

func (h *helloapp) AppURL(path string) string {
	conf := h.apps.Configurator.GetConfig()
	return conf.PluginURL + constants.HelloAppPath + path
}

type contextHandler func(http.ResponseWriter, *http.Request, *apps.JWTClaims, *api.Context) (int, error)
type callHandler func(http.ResponseWriter, *http.Request, *apps.JWTClaims, *api.Call) (int, error)
type notifyHandler func(http.ResponseWriter, *http.Request, *apps.JWTClaims, *api.Notification) (int, error)

func handleCall(r *mux.Router, path string, h callHandler) {
	r.HandleFunc(path,
		func(w http.ResponseWriter, req *http.Request) {
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
		},
	).Methods("POST")
}

func handleNotify(r *mux.Router, path string, h notifyHandler) {
	r.HandleFunc(path,
		func(w http.ResponseWriter, req *http.Request) {
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
		},
	).Methods("POST")
}

func handleGetWithContext(r *mux.Router, path string, h contextHandler) {
	r.HandleFunc(path,
		func(w http.ResponseWriter, req *http.Request) {
			claims, err := checkJWT(req)
			if err != nil {
				httputils.WriteBadRequestError(w, err)
				return
			}

			statusCode, err := h(w, req, claims, &api.Context{
				TeamID:       req.Form.Get(constants.TeamID),
				ChannelID:    req.Form.Get(constants.ChannelID),
				ActingUserID: req.Form.Get(constants.ActingUserID),
				PostID:       req.Form.Get(constants.PostID),
			})
			if err != nil {
				httputils.WriteJSONError(w, statusCode, "", err)
				return
			}
		},
	).Methods("GET")
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

func (h *helloapp) makeCall(path string, namevalues ...string) *api.Call {
	return api.MakeCall(h.AppURL(path), namevalues...)
}
