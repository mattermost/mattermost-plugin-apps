package helloapp

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

const (
	appID          = "hello"
	appDisplayName = "Hallo სამყარო"
	appDescription = "Hallo სამყარო test app"
)

func (h *helloapp) handleManifest(w http.ResponseWriter, req *http.Request) {
	httputils.WriteJSON(w,
		store.Manifest{
			AppID:       appID,
			DisplayName: appDisplayName,
			Description: appDescription,
			RootURL:     h.appURL(""),
			RequestedPermissions: []store.PermissionType{
				store.PermissionUserJoinedChannelNotification,
				store.PermissionActAsUser,
				store.PermissionActAsBot,
			},
			InstallFormURL:    h.appURL(pathInstall),
			OAuth2CallbackURL: h.appURL(pathOAuth2Complete),
			LocationsURL:      h.appURL(pathLocations),
			HomepageURL:       h.appURL("/"),
		})
}
