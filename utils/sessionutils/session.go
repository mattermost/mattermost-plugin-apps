package sessionutils

import (
	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func GetAppID(session *model.Session) apps.AppID {
	return apps.AppID(session.Props[model.SessionPropAppsFrameworkAppID])
}
