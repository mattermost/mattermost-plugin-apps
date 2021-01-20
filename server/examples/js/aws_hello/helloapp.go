package aws_hello

import "github.com/mattermost/mattermost-plugin-apps/server/api"

const (
	AppID          = "awsHello"
	AppDisplayName = "AWS Hello App display name"
	AppDescription = "AWS Hello App description"
	AppVersion     = "v0.0.1"
)

func Manifest() *api.Manifest {
	return &api.Manifest{
		AppID:       AppID,
		Version:     AppVersion,
		Type:        api.AppTypeAWSLambda,
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
		OnInstall: &api.Call{
			URL: "on_activate",
			Expand: &api.Expand{
				App:              api.ExpandAll,
				AdminAccessToken: api.ExpandAll,
			},
		},
	}
}
