package http_hello

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-api/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/examples/go/hello"
	"github.com/mattermost/mattermost-plugin-apps/server/appservices"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
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
	conf config.Service
	i18n *i18n.Bundle
}

// Init hello app router
func Init(router *mux.Router, mm *pluginapi.Client, conf config.Service, _ proxy.Service, _ appservices.Service, i18nBundle *i18n.Bundle) {
	h := helloapp{
		HelloApp: hello.NewHelloApp(mm),
		conf:     conf,
		i18n:     i18nBundle,
	}

	r := router.PathPrefix(config.HelloHTTPPath).Subrouter()
	r.HandleFunc(PathManifest, h.handleManifest).Methods("GET")

	handle(r, hello.PathInstall, h.Install)
	handle(r, apps.DefaultBindings.Path, h.GetBindings)
	handle(r, hello.PathSendSurvey+"/{type}", h.SendSurvey)
	handle(r, hello.PathSendSurveyModal+"/{type}", h.SendSurveyModal)
	handle(r, hello.PathSendSurveyCommandToModal+"/{type}", h.SendSurveyCommandToModal)
	handle(r, hello.PathSurvey+"/{type}", h.Survey)
	handle(r, hello.PathUserJoinedChannel+"/{type}", h.UserJoinedChannel)
	handle(r, hello.PathSubmitSurvey+"/{type}", h.SubmitSurvey)
}

func (h *helloapp) handleManifest(w http.ResponseWriter, req *http.Request) {
	httputils.WriteJSON(w,
		apps.Manifest{
			AppID:       AppID,
			AppType:     apps.AppTypeHTTP,
			DisplayName: AppDisplayName,
			Description: AppDescription,
			HTTPRootURL: h.appURL(""),
			RequestedPermissions: apps.Permissions{
				apps.PermissionUserJoinedChannelNotification,
				apps.PermissionActAsUser,
				apps.PermissionActAsBot,
				apps.PermissionActAsAdmin,
			},
			OnInstall: &apps.Call{
				Path: "/install",
				Expand: &apps.Expand{
					AdminAccessToken: apps.ExpandAll,
					App:              apps.ExpandAll,
				},
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

type handler func(http.ResponseWriter, *http.Request, *apps.JWTClaims, *apps.CallRequest) (int, error)

func handle(r *mux.Router, path string, h handler) {
	r.HandleFunc(path,
		func(w http.ResponseWriter, req *http.Request) {
			claims, err := checkJWT(req)
			if err != nil {
				proxy.WriteCallError(w, http.StatusUnauthorized, err)
				return
			}

			data, err := apps.CallRequestFromJSONReader(req.Body)
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
	conf := h.conf.GetConfig()
	return conf.PluginURL + config.HelloHTTPPath + path
}
