package aws_hello

import "github.com/mattermost/mattermost-plugin-apps/server/api"

const (
	AppID          = "my_app"
	AppDisplayName = "AWS Hello App display name"
	AppDescription = "AWS Hello App description"
)

func Manifest() *api.Manifest {
	return &api.Manifest{
		AppID:       AppID,
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
		Install: &api.Call{
			URL: "my_func",
			Expand: &api.Expand{
				App:              api.ExpandAll,
				AdminAccessToken: api.ExpandAll,
			},
		},
	}
}
