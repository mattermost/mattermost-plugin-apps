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
		AppType:     apps.AppTypeAWSLambda,
		DisplayName: AppDisplayName,
		Description: AppDescription,
		HomepageURL: ("https://github.com/mattermost"),

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

		OnInstall: &apps.Call{
			Path: "on_activate",
			Expand: &apps.Expand{
				App:              apps.ExpandAll,
				AdminAccessToken: apps.ExpandAll,
			},
		},
	}
}
