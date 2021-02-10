package aws_hello

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
)

const (
	AppID          = "awsHello"
	AppDisplayName = "AWS Hello App display name"
	AppDescription = "AWS Hello App description"
	AppVersion     = "v0.0.1"
)

func Manifest() *apps.Manifest {
	return &apps.Manifest{
		AppID:       AppID,
		Version:     AppVersion,
		Type:        apps.AppTypeAWSLambda,
		DisplayName: AppDisplayName,
		Description: AppDescription,
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
		HomepageURL: ("https://github.com/mattermost"),
		OnInstall: &apps.Call{
			URL: "on_activate",
			Expand: &apps.Expand{
				App:              apps.ExpandAll,
				AdminAccessToken: apps.ExpandAll,
			},
		},
	}
}
