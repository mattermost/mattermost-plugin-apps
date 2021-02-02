package aws_hello

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
)

const (
	AppID          = "awsHello"
	AppDisplayName = "AWS Hello App display name"
	AppDescription = "AWS Hello App description"
)

func Manifest() *apps.Manifest {
	return &apps.Manifest{
		AppID:       AppID,
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
		Install: &apps.Call{
			URL: "on_activate",
			Expand: &apps.Expand{
				App:              apps.ExpandAll,
				AdminAccessToken: apps.ExpandAll,
			},
		},
	}
}
