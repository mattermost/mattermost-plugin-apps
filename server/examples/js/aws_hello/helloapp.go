package aws_hello

import (
	"github.com/mattermost/mattermost-plugin-apps/modelapps"
)

const (
	AppID          = "awsHello"
	AppDisplayName = "AWS Hello App display name"
	AppDescription = "AWS Hello App description"
)

func Manifest() *modelapps.Manifest {
	return &modelapps.Manifest{
		AppID:       AppID,
		Type:        modelapps.AppTypeAWSLambda,
		DisplayName: AppDisplayName,
		Description: AppDescription,
		RequestedPermissions: modelapps.Permissions{
			modelapps.PermissionUserJoinedChannelNotification,
			modelapps.PermissionActAsUser,
			modelapps.PermissionActAsBot,
		},
		RequestedLocations: modelapps.Locations{
			modelapps.LocationChannelHeader,
			modelapps.LocationPostMenu,
			modelapps.LocationCommand,
			modelapps.LocationInPost,
		},
		HomepageURL: ("https://github.com/mattermost"),
		Install: &modelapps.Call{
			URL: "on_activate",
			Expand: &modelapps.Expand{
				App:              modelapps.ExpandAll,
				AdminAccessToken: modelapps.ExpandAll,
			},
		},
	}
}
