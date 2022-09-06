package appservices

import (
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-server/v6/model"
)

func (a *AppServices) RefreshBindings(r *incoming.Request, userID string) error {
	err := r.RequireSourceApp()
	if err != nil {
		return err
	}

	r.Config().MattermostAPI().Frontend.PublishWebSocketEvent(config.WebSocketEventRefreshBindings, map[string]interface{}{}, &model.WebsocketBroadcast{UserId: userID})
	return err
}
