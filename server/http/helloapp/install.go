package helloapp

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

func (h *helloapp) handleInstall(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims,
	data *apps.CallRequest) (int, error) {
	err := h.storeAppCredentials(&AppCredentials{
		BotAccessToken:     data.Values.Get("bot_access_token"),
		BotUserID:          data.Context.App.BotUserID,
		OAuth2ClientID:     data.Context.App.OAuth2ClientID,
		OAuth2ClientSecret: data.Values.Get("oauth2_client_secret"),
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
			Wish: &store.Wish{
				URL: h.AppURL(PathWishConnectedInstall),
			},
			Request: data,
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

func (h *helloapp) handleConnectedInstall(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, data *apps.CallRequest) (int, error) {
	var teams []*model.Team
	var team *model.Team
	var channel *model.Channel
	fmt.Printf("<><> ================ hello handleConnectedInstall 1:\n")

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
			fmt.Printf("<><> ================ hello handleConnectedInstall 2:\n")

			// TODO call a Modal to select a team
			team = teams[0]

			// Ensure "Hallo სამყარო" channel
			channel, _ = mmclient.GetChannelByName(AppID, team.Id, "")
			if channel != nil {
				// TODO DM to user that the channel has been found
				if channel.DeleteAt != 0 {
					return errors.Errorf("TODO unarchive channel %s \n", channel.DisplayName)
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
			_, api4Resp = mmclient.AddTeamMember(team.Id, data.Context.App.BotUserID)
			if api4Resp.Error != nil {
				return api4Resp.Error
			}
			_, api4Resp = mmclient.AddChannelMember(channel.Id, data.Context.App.BotUserID)
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

			// TODO this should be done using the REST Subs API, for now mock with direct use
			err = h.apps.Store.StoreSub(&store.Subscription{
				AppID:     AppID,
				Subject:   store.SubjectUserJoinedChannel,
				ChannelID: channel.Id,
				TeamID:    channel.TeamId,
				Expand: &store.Expand{
					Channel: store.ExpandAll,
					Team:    store.ExpandAll,
					User:    store.ExpandAll,
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

	ac, err := h.getAppCredentials()
	if err != nil {
		return http.StatusInternalServerError, err
	}
	httputils.WriteJSON(w,
		apps.CallResponse{
			Type:     apps.ResponseTypeOK,
			Markdown: md.Markdownf("installed %s (OAuth client ID: %s) to %s channel", AppDisplayName, ac.OAuth2ClientID, AppDisplayName),
		})
	h.DM(data.Context.ActingUserID, "OK!")

	return http.StatusOK, nil
}
