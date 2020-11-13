package helloapp

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

func (h *helloapp) handleManifest(w http.ResponseWriter, req *http.Request) {
	httputils.WriteJSON(w,
		apps.Manifest{
			AppID:       AppID,
			DisplayName: AppDisplayName,
			Description: AppDescription,
			RootURL:     h.appURL(""),
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
			OAuth2CallbackURL: h.appURL(PathOAuth2Complete),
			HomepageURL:       h.appURL("/"),
		})
}
