package http_hello

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/api/impl/proxy"
	"github.com/mattermost/mattermost-plugin-apps/server/examples/hello"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

const (
	AppID          = "http-hello"
	AppSecret      = "1234"
	AppDisplayName = "Hallo სამყარო (http)"
	AppDescription = "Hallo სამყარო HTTP test app"
)

const (
	PathManifest = "/mattermost-app.json"
)

type helloapp struct {
	*hello.HelloApp
}

// Init hello app router
func Init(router *mux.Router, appsService *api.Service) {
	h := helloapp{
		hello.NewHelloApp(appsService),
	}

	r := router.PathPrefix(api.HelloHTTPPath).Subrouter()
	r.HandleFunc(PathManifest, h.handleManifest).Methods("GET")

	handleCall(r, api.DefaultInstallCallPath, h.Install)
	handleCall(r, api.DefaultBindingsCallPath, h.GetBindings)

	handleCall(r, hello.PathSendSurvey, h.SendSurvey)
	handleCall(r, hello.PathSurvey, h.Survey)
	handleNotify(r, hello.PathUserJoinedChannel, h.UserJoinedChannel)
}

type callHandler func(http.ResponseWriter, *http.Request, *api.JWTClaims, *api.Call) (int, error)
type notifyHandler func(http.ResponseWriter, *http.Request, *api.JWTClaims, *api.Notification)

func handleCall(r *mux.Router, path string, h callHandler) {
	r.HandleFunc(path,
		func(w http.ResponseWriter, req *http.Request) {
			claims, err := checkJWT(req)
			if err != nil {
				proxy.WriteCallError(w, http.StatusUnauthorized, err)
				return
			}

			data, err := api.UnmarshalCallFromReader(req.Body)
			if err != nil {
				proxy.WriteCallError(w, http.StatusInternalServerError, err)
				return
			}

			status, err := h(w, req, claims, data)
			if err != nil && status != 0 && status != http.StatusOK {
				httputils.WriteJSONStatus(w, status, err)
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

			h(w, req, claims, &data)
		},
	).Methods("POST")
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
	conf := h.API.Configurator.GetConfig()
	return conf.PluginURL + api.HelloHTTPPath + path
}
