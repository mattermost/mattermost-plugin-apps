package hello

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/mmclient"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

func (h *HelloApp) Install(appID apps.AppID, channelDisplayName string, c *apps.Call) (md.MD, error) {
	if c.Type != apps.CallTypeSubmit {
		return "", errors.New("not supported")
	}

	bot := mmclient.AsBot(c.Context)
	adminClient := mmclient.AsAdmin(c.Context)

	var teams []*model.Team
	var team *model.Team
	var channel *model.Channel

	var api4Resp *model.Response
	teams, api4Resp = adminClient.GetAllTeams("", 0, 1)
	if api4Resp.Error != nil {
		return "", api4Resp.Error
	}
	if len(teams) == 0 {
		return "", errors.New("no team found to create the Hallo სამყარო channel")
	}

	// TODO call a Modal to select a team
	team = teams[0]

	// Ensure "Hallo სამყარო" channel
	channel, _ = adminClient.GetChannelByName(string(appID), team.Id, "")
	if channel != nil {
		// TODO DM to user that the channel has been found
		if channel.DeleteAt != 0 {
			return "", errors.Errorf("TODO unarchive channel %s \n", channel.DisplayName)
		}
		bot.DM(c.Context.ActingUserID, "Found existing ~%s channel.", appID)
	} else {
		channel, api4Resp = adminClient.CreateChannel(&model.Channel{
			TeamId:      team.Id,
			Name:        string(appID),
			DisplayName: channelDisplayName,
			Header:      "TODO header",
			Purpose:     `to say, "Hallo სამყარო!"`,
			Type:        model.CHANNEL_OPEN,
		})
		if api4Resp.Error != nil {
			return "", api4Resp.Error
		}

		bot.DM(c.Context.ActingUserID, "Created ~%s channel.", appID)
	}

	// Add the Bot user to the team and the channel.
	_, api4Resp = adminClient.AddTeamMember(team.Id, c.Context.App.BotUserID)
	if api4Resp.Error != nil {
		return "", api4Resp.Error
	}
	_, api4Resp = adminClient.AddChannelMember(channel.Id, c.Context.App.BotUserID)
	if api4Resp.Error != nil {
		return "", api4Resp.Error
	}

	bot.DM(c.Context.ActingUserID, "Added bot to channel.")

	_, _ = bot.CreatePost(&model.Post{
		ChannelId: channel.Id,
		Message:   fmt.Sprintf("%s has been installed into this channel and will now greet newly joining users", channelDisplayName),
	})
	bot.DM(c.Context.ActingUserID, "Posted welcome message to channel.")

	// TODO this should be done using the REST Subs API, for now mock with direct use
	_, err := adminClient.Subscribe(&apps.Subscription{
		AppID:     appID,
		Subject:   apps.SubjectUserJoinedChannel,
		ChannelID: channel.Id,
		TeamID:    channel.TeamId,
		Call: &apps.Call{
			Path: PathUserJoinedChannel,
			Expand: &apps.Expand{
				Channel: apps.ExpandAll,
				Team:    apps.ExpandAll,
				User:    apps.ExpandAll,
			},
		},
	})
	if err != nil {
		return "", err
	}
	bot.DM(c.Context.ActingUserID, "Subscribed to %s in channel.", apps.SubjectUserJoinedChannel)

	bot.DM(c.Context.ActingUserID, "Finished installing!")

	return md.Markdownf("installed %s to %s channel", appID, channelDisplayName), nil
}
