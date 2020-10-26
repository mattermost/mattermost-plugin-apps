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
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

const appSecret = "1234"

const (
	pathManifest                = "/mattermost-app.json"
	pathNotifyUserJoinedChannel = "/notify/" + string(store.SubjectUserJoinedChannel)

	pathInstall            = "/form/install"
	pathConnectedInstall   = "/form/connected_install"
	pathPing               = "/form/ping"
	pathCreateEmbeddedPing = "/form/embedded/ping"

	pathOAuth2         = "/oauth2"
	pathOAuth2Complete = "/oauth2/complete" // /complete comes from OAuther

	pathLocations = "/locations"

	pathDialogs          = "/dialog"
	pathOpenPingDialog   = pathDialogs + "/open/ping"
	pathSubmitPingDialog = pathDialogs + "/submit/ping"

	// DEBUG
	pathSubmitEmbedded = "/form/submit_embedded"
	pathCreateEmbedded = "/form/create_embedded"
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

	subrouter := router.PathPrefix(constants.HelloAppPath).Subrouter()

	subrouter.HandleFunc(pathManifest, h.handleManifest).Methods("GET")
	subrouter.PathPrefix(pathOAuth2).HandlerFunc(h.handleOAuth).Methods("GET")

	subrouter.HandleFunc(pathNotifyUserJoinedChannel, notify(h.handleUserJoinedChannel)).Methods("POST")

	subrouter.HandleFunc(pathInstall, call(h.handleInstall)).Methods("POST")
	subrouter.HandleFunc(pathConnectedInstall, call(h.handleConnectedInstall)).Methods("POST")

	subrouter.HandleFunc(pathPing, call(h.handlePing)).Methods("POST")
	subrouter.HandleFunc(pathCreateEmbeddedPing, call(h.handleCreatePingEmbedded)).Methods("POST")
	subrouter.HandleFunc(pathOpenPingDialog, call(h.handleOpenPingDialog)).Methods("POST")
	subrouter.HandleFunc(pathSubmitPingDialog, h.handleSubmitPingDialog).Methods("POST")

	subrouter.HandleFunc(pathLocations, checkAuthentication(extractUserAndChannelID(h.handleLocations))).Methods("GET")
	subrouter.HandleFunc(pathDialogs, checkAuthentication(h.handleDialog)).Methods("GET")

	// DEBUG
	subrouter.HandleFunc(pathSubmitEmbedded, call(h.handleSubmitEmbedded)).Methods("POST")
	subrouter.HandleFunc(pathCreateEmbedded, call(h.handleCreateEmbedded)).Methods("POST")

	_ = h.initOAuther()
}

func (h *helloapp) appURL(path string) string {
	conf := h.apps.Configurator.GetConfig()
	return conf.PluginURL + constants.HelloAppPath + path
}

type callHandler func(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, data *apps.Call) (int, error)

func call(h callHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		claims, err := checkJWT(req)
		if err != nil {
			httputils.WriteBadRequestError(w, err)
			return
		}

		data, err := apps.UnmarshalCallReader(req.Body)
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

type notifyHandler func(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, data *apps.Notification) (int, error)

func notify(h notifyHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
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
