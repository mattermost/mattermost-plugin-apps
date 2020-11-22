package builtin_hello

import (
	"github.com/mattermost/mattermost-plugin-apps/server/api"
)

func GetManifest() *api.Manifest {
	return &api.Manifest{
		AppID:       AppID,
		DisplayName: AppDisplayName,
		Description: AppDescription,
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
		HomepageURL: ("https://github.com/mattermost"),
	}
}
