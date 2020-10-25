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
	PathManifest                = "/mattermost-app.json"
	PathNotifyUserJoinedChannel = "/notify/" + string(api.SubjectUserJoinedChannel)
	PathInstall                 = "/f/install"
	PathConnectedInstall        = "/f/connected_install"
	PathSubscribe               = "/f/subscribe"
	PathMessage                 = "/f/message"
	PathHello                   = "/f/hello"
	PathOAuth2                  = "/oauth2"
	PathOAuth2Complete          = "/oauth2/complete" // /complete comes from OAuther
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

	handleFunction(subrouter, PathInstall, h.fInstall, h.fInstallMeta)
	handleFunction(subrouter, PathConnectedInstall, h.fConnectedInstall, nil)
	handleFunction(subrouter, PathMessage, h.fMessage, h.fMessageMeta)

	handleNotify(subrouter, PathNotifyUserJoinedChannel, h.handleUserJoinedChannel)

	_ = h.InitOAuther()
}

func (h *helloapp) handleManifest(w http.ResponseWriter, req *http.Request) {
	httputils.WriteJSON(w,
		api.Manifest{
			AppID:       AppID,
			DisplayName: AppDisplayName,
			Description: AppDescription,
			RootURL:     h.AppURL(""),
			RequestedPermissions: []api.PermissionType{
				api.PermissionUserJoinedChannelNotification,
				api.PermissionActAsUser,
				api.PermissionActAsBot,
			},
			Install:           h.AppURL(PathInstall),
			OAuth2CallbackURL: h.AppURL(PathOAuth2Complete),
			HomepageURL:       h.AppURL("/"),
		})
}

type fPostHandler func(http.ResponseWriter, *http.Request, *apps.JWTClaims, *api.Call) (int, error)
type fGetHandler func(http.ResponseWriter, *http.Request, *apps.JWTClaims, *api.Context) (int, error)
type nHandler func(http.ResponseWriter, *http.Request, *apps.JWTClaims, *api.Notification) (int, error)

func handleFunction(r *mux.Router, path string, ph fPostHandler, gh fGetHandler) {
	r.HandleFunc(path, fpost(ph)).Methods("POST")
	if gh != nil {
		r.HandleFunc(path, fget(gh)).Methods("GET")
	}
}

func handleNotify(r *mux.Router, path string, h nHandler) {
	r.HandleFunc(path, notify(h)).Methods("POST")
}

func fpost(h fPostHandler) http.HandlerFunc {
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

func fget(h fGetHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
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
	}
}

func notify(h nHandler) http.HandlerFunc {
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

func (h *helloapp) AppURL(path string) string {
	conf := h.apps.Configurator.GetConfig()
	return conf.PluginURL + constants.HelloAppPath + path
}

func (h *helloapp) makeCall(path string, namevalues ...string) *api.Call {
	c := &api.Call{
		URL: h.AppURL(path),
	}

	values := map[string]string{}
	for len(namevalues) > 0 {
		switch len(namevalues) {
		case 1:
			values[namevalues[0]] = ""
			namevalues = namevalues[1:]

		default:
			values[namevalues[0]] = values[namevalues[1]]
			namevalues = namevalues[2:]
		}
	}
	if len(values) > 0 {
		c.Values = values
	}
	return c
}
