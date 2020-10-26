package helloapp

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

const (
	AppID          = "hello"
	AppDisplayName = "Hallo სამყარო"
	AppDescription = "Hallo სამყარო test app"
)

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
			InstallFormURL:    h.AppURL(PathInstall),
			OAuth2CallbackURL: h.AppURL(PathOAuth2Complete),
			LocationsURL:      h.AppURL(PathLocations),
			HomepageURL:       h.AppURL("/"),
		})
}
