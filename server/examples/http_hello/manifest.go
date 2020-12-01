package http_hello

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

func (h *helloapp) handleManifest(w http.ResponseWriter, req *http.Request) {
	httputils.WriteJSON(w,
		api.Manifest{
			AppID:         AppID,
			DisplayName:   AppDisplayName,
			Description:   AppDescription,
			RootURL: h.appURL(""),
			RequestedPermissions: api.Permissions{
				api.PermissionUserJoinedChannelNotification,
				api.PermissionActAsUser,
				api.PermissionActAsBot,
			},
			RequestedLocations: api.Locations{
				api.LocationChannelHeader,
				api.LocationPostMenu,
				api.LocationCommand,
				api.LocationInPost,
			},
			HomepageURL: h.appURL("/"),
		})
}
