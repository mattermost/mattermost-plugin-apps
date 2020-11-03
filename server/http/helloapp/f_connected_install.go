package helloapp

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

func (h *helloapp) fConnectedInstall(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, call *api.Call) (int, error) {
	var teams []*model.Team
	var team *model.Team
	var channel *model.Channel

	err := h.asUser(call.Context.ActingUserID,
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
			channel, _ = mmclient.GetChannelByName(appID, team.Id, "")
			if channel != nil {
				// TODO DM to user that the channel has been found
				if channel.DeleteAt != 0 {
					return errors.Errorf("TODO unarchive channel %s \n", channel.DisplayName)
				}
				h.dm(call.Context.ActingUserID, "Found existing ~%s channel.", appID)
			} else {
				channel, api4Resp = mmclient.CreateChannel(&model.Channel{
					TeamId:      team.Id,
					Type:        model.CHANNEL_OPEN,
					DisplayName: appDisplayName,
					Name:        appID,
					Header:      "TODO header",
					Purpose:     `to say, "Hallo სამყარო!"`,
				})
				if api4Resp.Error != nil {
					return api4Resp.Error
				}

				h.dm(call.Context.ActingUserID, "Created ~%s channel.", appID)
			}

			// Add the Bot user to the team and the channel.
			_, api4Resp = mmclient.AddTeamMember(team.Id, call.Context.App.BotUserID)
			if api4Resp.Error != nil {
				return api4Resp.Error
			}
			_, api4Resp = mmclient.AddChannelMember(channel.Id, call.Context.App.BotUserID)
			if api4Resp.Error != nil {
				return api4Resp.Error
			}

			h.dm(call.Context.ActingUserID, "Added bot to channel.")
			return nil
		})
	if err != nil {
		return http.StatusInternalServerError, err
	}

	err = h.asBot(
		func(mmclient *model.Client4, botUserID string) error {
			_, _ = mmclient.CreatePost(&model.Post{
				ChannelId: channel.Id,
				Message:   fmt.Sprintf("%s has been installed into this channel and will now greet newly joining users", appDisplayName),
			})
			h.dm(call.Context.ActingUserID, "Posted welcome message to channel.")

			// TODO this should be done using the REST Subs API, for now mock with direct use
			err = h.apps.Store.StoreSub(&api.Subscription{
				AppID:     appID,
				Subject:   api.SubjectUserJoinedChannel,
				ChannelID: channel.Id,
				TeamID:    channel.TeamId,
				Expand: &api.Expand{
					Channel: api.ExpandAll,
					Team:    api.ExpandAll,
					User:    api.ExpandAll,
				},
			})
			if err != nil {
				return err
			}
			h.dm(call.Context.ActingUserID, "Subscribed to %s in channel.", api.SubjectUserJoinedChannel)
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
		api.CallResponse{
			Type:     api.CallResponseTypeOK,
			Markdown: md.Markdownf("installed %s (OAuth client ID: %s) to %s channel", appDisplayName, ac.OAuth2ClientID, appDisplayName),
		})
	h.dm(call.Context.ActingUserID, "OK!")

	return http.StatusOK, nil
}
