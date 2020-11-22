package http_hello

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
	PathInstall        = api.AppInstallPath  // convention for Mattermost Apps
	PathBindings       = api.AppBindingsPath // convention for Mattermost Apps
	PathOAuth2         = "/oauth2"           // convention for Mattermost Apps, comes from OAuther
	PathOAuth2Complete = "/oauth2/complete"  // convention for Mattermost Apps, comes from OAuther

	PathConnectedInstall = "/connected_install"
	PathSendSurvey       = "/send"
	PathSubscribeChannel = "/subscribe"
	PathSurvey           = "/survey"

	PathNotifyUserJoinedChannel = "/notify-user-joined-channel"
)

type helloapp struct {
	api     *api.Service
	OAuther oauther.OAuther
}

// Init hello app router
func Init(router *mux.Router, appsService *api.Service) {
	h := helloapp{
		api: appsService,
	}

	r := router.PathPrefix(api.HelloHTTPPath).Subrouter()
	r.HandleFunc(PathManifest, h.handleManifest).Methods("GET")
	handleGetWithContext(r, PathBindings, h.bindings)
	r.PathPrefix(PathOAuth2).HandlerFunc(h.handleOAuth).Methods("GET")

	// Naming convention: fXXX are "Callable" functions, nXXX are notification
	// handlers.
	handleCall(r, PathInstall, h.fInstall)
	handleCall(r, PathConnectedInstall, h.fConnectedInstall)
	handleCall(r, PathSendSurvey, h.fSendSurvey)
	handleCall(r, PathSurvey, h.fSurvey)

	handleNotify(r, PathNotifyUserJoinedChannel, h.nUserJoinedChannel)

	_ = h.initOAuther()
}

type contextHandler func(http.ResponseWriter, *http.Request, *api.JWTClaims, *api.Context) (int, error)
type callHandler func(http.ResponseWriter, *http.Request, *api.JWTClaims, *api.Call) (int, error)
type notifyHandler func(http.ResponseWriter, *http.Request, *api.JWTClaims, *api.Notification) (int, error)

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
				TeamID:       req.Form.Get(api.PropTeamID),
				ChannelID:    req.Form.Get(api.PropChannelID),
				ActingUserID: req.Form.Get(api.PropActingUserID),
				PostID:       req.Form.Get(api.PropPostID),
			})
			if err != nil {
				httputils.WriteJSONError(w, statusCode, "", err)
				return
			}
		},
	).Methods("GET")
}

func checkJWT(req *http.Request) (*api.JWTClaims, error) {
	authValue := req.Header.Get(api.OutgoingAuthHeader)
	if !strings.HasPrefix(authValue, "Bearer ") {
		return nil, errors.Errorf("missing %s: Bearer header", api.OutgoingAuthHeader)
	}

	jwtoken := strings.TrimPrefix(authValue, "Bearer ")
	claims := api.JWTClaims{}
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
	conf := h.api.Configurator.GetConfig()
	return conf.PluginURL + api.HelloHTTPPath + path
}

func (h *helloapp) makeCall(path string, namevalues ...string) *api.Call {
	return api.MakeCall(h.appURL(path), namevalues...)
}
