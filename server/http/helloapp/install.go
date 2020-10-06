package helloapp

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

func (h *helloapp) handleInstall(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, data *apps.CallData) (int, error) {
	err := h.storeAppCredentials(
		data.Expanded.App.OAuthAppID,
		data.Expanded.App.OAuthSecret,
		data.Expanded.App.BotUserID,
		data.Expanded.App.BotPersonalAccessToken)
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
	var teams []*model.Team
	var team *model.Team
	var channel *model.Channel

	err := h.asUser(data.Context.ActingUserID,
		func(mmclient *model.Client4) error {
			var api4Resp *model.Response
			teams, api4Resp = mmclient.GetAllTeams("", 0, 1)
			if api4Resp.Error != nil {
				return api4Resp.Error
			}
			if len(teams) == 0 {
				return errors.New("no team found to create the Hallo სამყარო channel")
			}

			// TODO call a Modal to select a team
			team = teams[0]

			// Ensure "Hallo სამყარო" channel
			channel, _ = mmclient.GetChannelByName(AppID, team.Id, "")
			if channel != nil {
				// TODO DM to user that the channel has been found
				if channel.DeleteAt != 0 {
					return errors.Errorf("<><> TODO unarchive channel %s \n", channel.DisplayName)
				}
			} else {
				channel, api4Resp = mmclient.CreateChannel(&model.Channel{
					TeamId:      team.Id,
					Type:        model.CHANNEL_OPEN,
					DisplayName: AppDisplayName,
					Name:        AppID,
					Header:      "TODO header",
					Purpose:     `to say, "Hallo სამყარო!"`,
				})
				if api4Resp.Error != nil {
					return api4Resp.Error
				}
				// TODO DM to user that the channel has been created
			}

			// Add the Bot user to the team and the channel.
			_, api4Resp = mmclient.AddTeamMember(team.Id, data.Expanded.App.BotUserID)
			if api4Resp.Error != nil {
				return api4Resp.Error
			}
			_, api4Resp = mmclient.AddChannelMember(channel.Id, data.Expanded.App.BotUserID)
			if api4Resp.Error != nil {
				return api4Resp.Error
			}

			return nil
		})
	if err != nil {
		return http.StatusInternalServerError, err
	}

	err = h.asBot(
		func(mmclient *model.Client4, botUserID string) error {
			_, _ = mmclient.CreatePost(&model.Post{
				ChannelId: channel.Id,
				Message:   fmt.Sprintf("%s has been installed into this channel and will now greet newly joining users", AppDisplayName),
			})

			// TODO Create a Subscription to UserJoinedChannel, handler
			return nil
		})
	if err != nil {
		return http.StatusInternalServerError, err
	}

	ac, err := h.getAppCredentials()
	if err != nil {
		return http.StatusInternalServerError, err
	}
	httputils.WriteJSON(w,
		apps.CallResponse{
			Type:     apps.ResponseTypeOK,
			Markdown: md.Markdownf("installed %s (OAuth client ID: %s) to %s channel", AppDisplayName, ac.OAuth2AppID, AppDisplayName),
		})
	return http.StatusOK, nil
}
