package helloapp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/mattermost/mattermost-plugin-apps/server/configurator"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"

	"github.com/gorilla/mux"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/constants"
)

const AppSecret = "1234"

type helloapp struct {
	mm           *pluginapi.Client
	configurator configurator.Service
}

func Init(router *mux.Router, apps *apps.Service) {
	a := helloapp{
		mm:           apps.Mattermost,
		configurator: apps.Configurator,
	}

	subrouter := router.PathPrefix(constants.HelloAppPath).Subrouter()
	subrouter.HandleFunc("/mattermost-app.json", a.handleManifest).Methods("GET")

	subrouter.HandleFunc("/wish/install", a.handleInstall).Methods("POST")
}

func (h *helloapp) handleManifest(w http.ResponseWriter, req *http.Request) {
	conf := h.configurator.GetConfig()

	rootURL := conf.PluginURL + constants.HelloAppPath

	httputils.WriteJSON(w,
		apps.Manifest{
			AppID:       "hello",
			DisplayName: "Hallo სამყარო",
			Description: "Hallo სამყარო test app",
			RootURL:     rootURL,
			RequestedPermissions: []apps.PermissionType{
				apps.PermissionUserJoinedChannelNotification,
				apps.PermissionActAsUser,
				apps.PermissionActAsBot,
			},
			Install: &apps.Wish{
				URL: rootURL + "/wish/install",
			},
		})
}

func (h *helloapp) handleInstall(w http.ResponseWriter, req *http.Request) {
	authValue := req.Header.Get(apps.AuthHeader)
	if !strings.HasPrefix(authValue, "Bearer ") {
		httputils.WriteBadRequestError(w, errors.Errorf("missing %s: Bearer header", apps.AuthHeader))
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

	conf := h.configurator.GetConfig()
	mmClient := model.NewAPIv4Client(conf.MattermostSiteURL)

	// TODO get the token from the request
	mmClient.SetToken("kuk76mfjhpnppruqpmkkeuhacw")

	logChannelID, logRootPostID := "", ""
	if data.Env != nil {
		v, ok := data.Env["log_channel_id"]
		if ok {
			logChannelID = v.(string)
		}

		v, ok = data.Env["log_root_post_id"]
		if ok {
			logRootPostID = v.(string)
		}
	}
	logDM := func(m string) {
		if logChannelID == "" {
			return
		}
		_, _ = mmClient.CreatePost(&model.Post{
			UserId:    conf.BotUserID,
			ChannelId: logChannelID,
			RootId:    logRootPostID,
			ParentId:  logRootPostID,
			Message:   m,
			Type:      model.POST_DEFAULT,
		})
	}

	teams, _ := mmClient.GetAllTeams("", 0, 100)
	if len(teams) == 0 {
		httputils.WriteJSONStatus(w, http.StatusInternalServerError,
			apps.CallResponse{
				Type:  apps.ResponseTypeError,
				Error: errors.New("no teams found to create Hallo სამყარო channel in"),
			})
		return
	}

	channel, api4Resp := mmClient.CreateChannel(&model.Channel{
		TeamId:      teams[0].Id,
		Type:        model.CHANNEL_OPEN,
		DisplayName: "Hallo სამყარო",
		Name:        "hello",
		Header:      "Hallo სამყარო header",
		Purpose:     "inquires about new member's emotional state",
	})
	if channel == nil {
		httputils.WriteJSONStatus(w, http.StatusInternalServerError,
			apps.CallResponse{
				Type:  apps.ResponseTypeError,
				Error: errors.Wrapf(api4Resp.Error, "failed to create ~Hallo სამყარო channel in %v\n", teams[0].Id),
			})
		return
	}

	_, _ = mmClient.CreatePost(&model.Post{
		UserId:    conf.BotUserID,
		ChannelId: channel.Id,
		Message:   "Users joining this channel will be asked about their well-being, and the information displayed publicly.",
		Type:      model.POST_DEFAULT,
	})
	logDM("created ~Hallo სამყარო")

	// TODO Subscribe

	httputils.WriteJSON(w,
		apps.CallResponse{
			Type:     apps.ResponseTypeOK,
			Markdown: "Installed! <><>",
			Data:     map[string]interface{}{"status": "ok"},
		})
}
