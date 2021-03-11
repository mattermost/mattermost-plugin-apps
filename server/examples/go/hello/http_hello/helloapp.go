package http_hello

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/api/impl/proxy"
	"github.com/mattermost/mattermost-plugin-apps/server/examples/go/hello"
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

	handle(r, apps.DefaultInstallCallPath, h.Install)
	handle(r, apps.DefaultBindingsCallPath, h.GetBindings)
	handle(r, hello.PathSendSurvey, h.SendSurvey)
	handle(r, hello.PathSendSurveyModal, h.SendSurveyModal)
	handle(r, hello.PathSendSurveyCommandToModal, h.SendSurveyCommandToModal)
	handle(r, hello.PathSurvey, h.Survey)
	handle(r, hello.PathUserJoinedChannel, h.UserJoinedChannel)
	handle(r, hello.PathSubmitSurvey, h.SubmitSurvey)
}

func (h *helloapp) handleManifest(w http.ResponseWriter, req *http.Request) {
	httputils.WriteJSON(w,
		apps.Manifest{
			AppID:       AppID,
			Type:        apps.AppTypeHTTP,
			DisplayName: AppDisplayName,
			Description: AppDescription,
			HTTPRootURL: h.appURL(""),
			RequestedPermissions: apps.Permissions{
				apps.PermissionUserJoinedChannelNotification,
				apps.PermissionActAsUser,
				apps.PermissionActAsBot,
			},
			RequestedLocations: apps.Locations{
				apps.LocationChannelHeader,
				apps.LocationPostMenu,
				apps.LocationCommand,
				apps.LocationInPost,
			},
			HomepageURL: h.appURL("/"),
		})
}

type handler func(http.ResponseWriter, *http.Request, *api.JWTClaims, *apps.CallRequest) (int, error)

func handle(r *mux.Router, path string, h handler) {
	r.HandleFunc(path,
		func(w http.ResponseWriter, req *http.Request) {
			claims, err := checkJWT(req)
			if err != nil {
				proxy.WriteCallError(w, http.StatusUnauthorized, err)
				return
			}

			data, err := apps.UnmarshalCallRequestFromReader(req.Body)
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
