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

func (h *helloapp) fConnectedInstall(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, c *apps.Call) (int, error) {
	if c.Type != apps.CallTypeSubmit {
		return http.StatusBadRequest, errors.New("not supported")
	}

	var teams []*model.Team
	var team *model.Team
	var channel *model.Channel

	err := h.asUser(c.Context.ActingUserID,
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
					return errors.Errorf("TODO unarchive channel %s \n", channel.DisplayName)
				}
				h.dm(c.Context.ActingUserID, "Found existing ~%s channel.", AppID)
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

				h.dm(c.Context.ActingUserID, "Created ~%s channel.", AppID)
			}

			// Add the Bot user to the team and the channel.
			_, api4Resp = mmclient.AddTeamMember(team.Id, c.Context.App.BotUserID)
			if api4Resp.Error != nil {
				return api4Resp.Error
			}
			_, api4Resp = mmclient.AddChannelMember(channel.Id, c.Context.App.BotUserID)
			if api4Resp.Error != nil {
				return api4Resp.Error
			}

			h.dm(c.Context.ActingUserID, "Added bot to channel.")
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
			h.dm(c.Context.ActingUserID, "Posted welcome message to channel.")

			// TODO this should be done using the REST Subs API, for now mock with direct use
			err = h.apps.API.Subscribe(&apps.Subscription{
				AppID:     AppID,
				Subject:   apps.SubjectUserJoinedChannel,
				ChannelID: channel.Id,
				TeamID:    channel.TeamId,
				Expand: &apps.Expand{
					Channel: apps.ExpandAll,
					Team:    apps.ExpandAll,
					User:    apps.ExpandAll,
				},
			})
			if err != nil {
				return err
			}
			h.dm(c.Context.ActingUserID, "Subscribed to %s in channel.", apps.SubjectUserJoinedChannel)
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
			Type:     apps.CallResponseTypeOK,
			Markdown: md.Markdownf("installed %s (OAuth client ID: %s) to %s channel", AppDisplayName, ac.OAuth2ClientID, AppDisplayName),
		})
	h.dm(c.Context.ActingUserID, "OK!")

	return http.StatusOK, nil
}
