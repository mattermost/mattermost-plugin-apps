package helloapp

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

const (
	AppID          = "hello"
	AppDisplayName = "Hallo სამყარო"
	AppDescription = "Hallo სამყარო test app"
)

func (h *helloapp) handleManifest(w http.ResponseWriter, req *http.Request) {
	httputils.WriteJSON(w,
		store.Manifest{
			AppID:       AppID,
			DisplayName: AppDisplayName,
			Description: AppDescription,
			RootURL:     h.AppURL(""),
			RequestedPermissions: []store.PermissionType{
				store.PermissionUserJoinedChannelNotification,
				store.PermissionActAsUser,
				store.PermissionActAsBot,
			},
			InstallFormURL:    h.AppURL(PathFormInstall),
			OAuth2CallbackURL: h.AppURL(PathOAuth2Complete),
			HomepageURL:       h.AppURL("/"),
		})
}
