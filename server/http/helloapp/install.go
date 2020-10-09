package helloapp

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/constants"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
	"github.com/mattermost/mattermost-server/v5/model"
)

func (h *helloapp) handleInstall(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, data *apps.CallData) (int, error) {
	err := h.storeAppCredentials(&AppCredentials{
		BotAccessToken:     data.Expanded.App.BotPersonalAccessToken,
		BotUserID:          data.Expanded.App.BotUserID,
		OAuth2ClientID:     data.Expanded.App.OAuthAppID,
		OAuth2ClientSecret: data.Expanded.App.OAuthSecret,
	})
	if err != nil {
		return http.StatusInternalServerError, err
	}
	err = h.InitOAuther()
	if err != nil {
		return http.StatusInternalServerError, err
	}

	connectURL, err := h.startOAuth2Connect(
		data.Context.ActingUserID,
		apps.Call{
			Wish: apps.NewWish(h.AppURL(PathWishConnectedInstall)),
			Data: data,
		})
	if err != nil {
		return http.StatusInternalServerError, err
	}

	httputils.WriteJSON(w,
		apps.CallResponse{
			Type: apps.ResponseTypeOK,
			Markdown: md.Markdownf(
				"**Hallo სამყარო** needs to continue its installation using your system administrator's credentials. Please [connect](%s) the application to your Mattermost account.",
				connectURL),
		})
	return http.StatusOK, nil
}

func (h *helloapp) handleConnectedInstall(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, data *apps.CallData) (int, error) {
	var channel *model.Channel
	err := h.asUser(data.Context.ActingUserID,
		func(client *model.Client4) error {
			channel, _ = client.GetChannelByName(AppID, data.Context.TeamID, "")
			if channel == nil {
				var api4Resp *model.Response
				channel, api4Resp = client.CreateChannel(
					&model.Channel{
						TeamId:      data.Context.TeamID,
						Type:        model.CHANNEL_OPEN,
						DisplayName: AppDisplayName,
						Name:        AppID,
					})
				if api4Resp.Error != nil {
					return api4Resp.Error
				}
			}
			return nil
		})
	if err != nil {
		return http.StatusInternalServerError, err
	}

	err = h.asBot(
		func(client *model.Client4) error {
			// TODO this should be done using the REST Subs API, for now mock with direct use
			err := h.apps.Subscriptions.StoreSub(apps.Subscription{
				SubscriptionID: apps.SubscriptionID(model.NewId()),
				AppID:          AppID,
				Subject:        constants.SubjectUserJoinedChannel,
				ChannelID:      channel.Id,
				TeamID:         channel.TeamId,
				Expand: &apps.Expand{
					Channel: apps.ExpandAll,
					Team:    apps.ExpandAll,
					User:    apps.ExpandAll,
				},
			})
			if err != nil {
				return err
			}
			return nil
		})
	if err != nil {
		return http.StatusInternalServerError, err
	}

	h.DM(data.Context.ActingUserID, "OK!")

	httputils.WriteJSON(w,
		apps.CallResponse{
			Type:     apps.ResponseTypeOK,
			Markdown: md.Markdownf("OK"),
		})
	return http.StatusOK, nil
}
