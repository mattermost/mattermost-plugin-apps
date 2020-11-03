package helloapp

import (
	"fmt"
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
	man := store.Manifest{
		AppID:       AppID,
		DisplayName: AppDisplayName,
		Description: AppDescription,
		RootURL:     h.AppURL(""),
		RequestedPermissions: []store.PermissionType{
			store.PermissionUserJoinedChannelNotification,
			store.PermissionActAsUser,
			store.PermissionActAsBot,
		},
		// InstallFormURL: h.AppURL(PathInstall),
		InstallFormURL:    "https://znd374xbk6.execute-api.us-east-2.amazonaws.com/dev",
		OAuth2CallbackURL: "https://m20ldarqw9.execute-api.us-east-2.amazonaws.com/dev",
		LocationsURL:      h.AppURL(PathLocations),
		HomepageURL:       h.AppURL("/"),
	}
	println(fmt.Sprintf("man = %v", man))

	httputils.WriteJSON(w,
		man)
}
