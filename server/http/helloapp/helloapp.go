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
	PathInstall        = apps.AppInstallPath  // convention for Mattermost Apps
	PathBindings       = apps.AppBindingsPath // convention for Mattermost Apps
	PathOAuth2         = "/oauth2"            // convention for Mattermost Apps, comes from OAuther
	PathOAuth2Complete = "/oauth2/complete"   // convention for Mattermost Apps, comes from OAuther

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
func Init(router *mux.Router, appsService *apps.Service) {
	h := helloapp{
		apps: appsService,
	}

	r := router.PathPrefix(apps.HelloAppPath).Subrouter()
	r.HandleFunc(PathManifest, h.handleManifest).Methods("GET")
	handleGetWithContext(r, PathBindings, h.bindings)
	r.PathPrefix(PathOAuth2).HandlerFunc(h.handleOAuth).Methods("GET")

	handleCall(r, PathInstall, h.fInstall)
	handleCall(r, PathConnectedInstall, h.fConnectedInstall)
	handleCall(r, PathSendSurvey, h.fSendSurvey)
	handleCall(r, PathSurvey, h.fSurvey)

	handleNotify(r, PathNotifyUserJoinedChannel, h.nUserJoinedChannel)

	_ = h.initOAuther()
}

type contextHandler func(http.ResponseWriter, *http.Request, *apps.JWTClaims, *apps.Context) (int, error)
type callHandler func(http.ResponseWriter, *http.Request, *apps.JWTClaims, *apps.Call) (int, error)
type notifyHandler func(http.ResponseWriter, *http.Request, *apps.JWTClaims, *apps.Notification) (int, error)

func handleCall(r *mux.Router, path string, h callHandler) {
	r.HandleFunc(path,
		func(w http.ResponseWriter, req *http.Request) {
			claims, err := checkJWT(req)
			if err != nil {
				httputils.WriteBadRequestError(w, err)
				return
			}

			data, err := apps.UnmarshalCallFromReader(req.Body)
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

			data := apps.Notification{}
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

			statusCode, err := h(w, req, claims, &apps.Context{
				TeamID:       req.Form.Get(apps.PropTeamID),
				ChannelID:    req.Form.Get(apps.PropChannelID),
				ActingUserID: req.Form.Get(apps.PropActingUserID),
				PostID:       req.Form.Get(apps.PropPostID),
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

func (h *helloapp) appURL(path string) string {
	conf := h.apps.Configurator.GetConfig()
	return conf.PluginURL + apps.HelloAppPath + path
}

func (h *helloapp) makeCall(path string, namevalues ...string) *apps.Call {
	return apps.MakeCall(h.appURL(path), namevalues...)
}
