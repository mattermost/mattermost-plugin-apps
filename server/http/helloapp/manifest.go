package helloapp

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

func (h *helloapp) handleManifest(w http.ResponseWriter, req *http.Request) {
	httputils.WriteJSON(w,
		api.Manifest{
			AppID:       appID,
			DisplayName: appDisplayName,
			Description: appDescription,
			RootURL:     h.appURL(""),
			RequestedPermissions: []api.PermissionType{
				api.PermissionUserJoinedChannelNotification,
				api.PermissionActAsUser,
				api.PermissionActAsBot,
			},
			OAuth2CallbackURL: h.appURL(pathOAuth2Complete),
			HomepageURL:       h.appURL("/"),
		})
}
