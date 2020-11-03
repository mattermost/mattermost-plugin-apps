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
	appID          = "hello"
	appSecret      = "1234"
	appDisplayName = "Hallo სამყარო"
	appDescription = "Hallo სამყარო test app"
)

const (
	pathManifest       = "/mattermost-app.json"
	pathInstall        = constants.AppInstallPath  // convention for Mattermost Apps
	pathBindings       = constants.AppBindingsPath // convention for Mattermost Apps
	pathOAuth2         = "/oauth2"                 // convention for Mattermost Apps, comes from OAuther
	pathOAuth2Complete = "/oauth2/complete"        // convention for Mattermost Apps, comes from OAuther
	pathDebugDialogs   = "/dialog"

	pathConnectedInstall = "/connected_install"
	pathSendMessage      = "/message"

	pathCreateEmbedded = "/create_embedded"
	// pathCreateEmbeddedPing      = "/embedded/ping"
	// pathHello                   = "/hello"
	pathNotifyUserJoinedChannel = "/notify-user-joined-channel"
	// pathOpenPingDialog          = pathDebugDialogs + "/open/ping"
	pathPing             = "/ping"
	pathSubmitEmbedded   = "/submit_embedded"
	pathSubmitPingDialog = pathDebugDialogs + "/submit/ping"
	pathSubscribeChannel = "/subscribe"
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
	// <<<<<<< HEAD

	// 	subrouter.HandleFunc(pathManifest, h.handleManifest).Methods("GET")
	// 	subrouter.PathPrefix(pathOAuth2).HandlerFunc(h.handleOAuth).Methods("GET")

	// 	subrouter.HandleFunc(pathNotifyUserJoinedChannel, notify(h.handleUserJoinedChannel)).Methods("POST")

	// 	subrouter.HandleFunc(pathInstall, call(h.handleInstall)).Methods("POST")
	// 	subrouter.HandleFunc(pathConnectedInstall, call(h.handleConnectedInstall)).Methods("POST")

	// 	subrouter.HandleFunc(pathPing, call(h.handlePing)).Methods("POST")
	// 	subrouter.HandleFunc(pathSubmitPingDialog, h.handleSubmitPingDialog).Methods("POST")

	// 	subrouter.HandleFunc(pathLocations, checkAuthentication(extractUserAndChannelID(h.handleLocations))).Methods("GET")
	// 	subrouter.HandleFunc(pathDialogs, checkAuthentication(h.handleDialog)).Methods("GET")

	// 	// DEBUG
	// 	subrouter.HandleFunc(pathSubmitEmbedded, call(h.handleSubmitEmbedded)).Methods("POST")
	// 	subrouter.HandleFunc(pathCreateEmbedded, call(h.handleCreateEmbedded)).Methods("POST")
	// =======
	r.HandleFunc(pathManifest, h.handleManifest).Methods("GET")
	r.PathPrefix(pathOAuth2).HandlerFunc(h.handleOAuth).Methods("GET")

	handleGetWithContext(r, pathBindings, (h.handleBindings))

	handleCall(r, pathInstall, h.Install)
	handleCall(r, pathConnectedInstall, h.ConnectedInstall)
	handleCall(r, pathSendMessage, h.Message)

	// handleCall(r, pathSubmitEmbedded, h.handleSubmitEmbedded)
	// handleCall(r, pathCreateEmbedded, h.handleCreateEmbedded)
	// handleCall(r, pathCreateEmbeddedPing, h.handleCreatePingEmbedded)
	// r.HandleFunc(pathOpenPingDialog, call(h.handleOpenPingDialog)).Methods("POST")

	handleNotify(r, pathNotifyUserJoinedChannel, h.handleUserJoinedChannel)

	_ = h.initOAuther()
}

func (h *helloapp) appURL(path string) string {
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
		return []byte(appSecret), nil
	})
	if err != nil {
		return nil, err
	}

	return &claims, nil
}

func (h *helloapp) handleDialog(w http.ResponseWriter, req *http.Request, _ apps.JWTClaims) {
	dialogID := req.URL.Query().Get("dialogID")
	if dialogID == "" {
		httputils.WriteBadRequestError(w, errors.New("dialog id not provided"))
		return
	}

	dialog, err := h.getDialog(dialogID)
	if err != nil {
		httputils.WriteInternalServerError(w, errors.Wrap(err, "error while getting dialog"))
		return
	}

	httputils.WriteJSON(w, dialog)
}

func (h *helloapp) makeCall(path string, namevalues ...string) *api.Call {
	return api.MakeCall(h.appURL(path), namevalues...)
}
