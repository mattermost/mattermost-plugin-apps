package builtin_hello

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/examples"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
	"github.com/mattermost/mattermost-server/v5/model"
)

func Install(c *api.Call) *api.CallResponse {
	if c.Type != api.CallTypeSubmit {
		return callError(errors.New("not supported"))
	}

	var teams []*model.Team
	var team *model.Team
	var channel *model.Channel

	asAdmin := examples.AsAdmin(c.Context)
	asBot := examples.AsBot(c.Context)

	var api4Resp *model.Response
	teams, api4Resp = asAdmin.GetAllTeams("", 0, 1)
	if api4Resp.Error != nil {
		return callError(api4Resp.Error)
	}
	if len(teams) == 0 {
		return callError(errors.New("no team found to create the Hallo სამყარო channel"))
	}

	// TODO call a Modal to select a team
	team = teams[0]

	// Ensure "Hallo სამყარო" channel
	channel, _ = asAdmin.GetChannelByName(AppID, team.Id, "")
	if channel != nil {
		// TODO DM to user that the channel has been found
		if channel.DeleteAt != 0 {
			return callError(errors.Errorf("TODO unarchive channel %s \n", channel.DisplayName))
		}
		asBot.DM(c.Context.ActingUserID, "Found existing ~%s channel.", AppID)
	} else {
		channel, api4Resp = asAdmin.CreateChannel(&model.Channel{
			TeamId:      team.Id,
			Type:        model.CHANNEL_OPEN,
			DisplayName: AppDisplayName,
			Name:        AppID,
			Header:      "TODO header",
			Purpose:     `to say, "Hallo სამყარო!"`,
		})
		if api4Resp.Error != nil {
			return callError(api4Resp.Error)
		}

		asBot.DM(c.Context.ActingUserID, "Created ~%s channel.", AppID)
	}

	// Add the Bot user to the team and the channel.
	_, api4Resp = asAdmin.AddTeamMember(team.Id, c.Context.App.BotUserID)
	if api4Resp.Error != nil {
		return callError(api4Resp.Error)
	}
	_, api4Resp = asAdmin.AddChannelMember(channel.Id, c.Context.App.BotUserID)
	if api4Resp.Error != nil {
		return callError(api4Resp.Error)
	}
	asBot.DM(c.Context.ActingUserID, "Added bot to channel.")

	_, _ = asBot.CreatePost(&model.Post{
		ChannelId: channel.Id,
		Message:   fmt.Sprintf("%s has been installed into this channel and will now greet newly joining users", AppDisplayName),
	})
	asBot.DM(c.Context.ActingUserID, "Posted welcome message to channel.")

	// TODO subscribe using the REST API
	// &api.Subscription{
	// 	AppID:     AppID,
	// 	Subject:   api.SubjectUserJoinedChannel,
	// 	ChannelID: channel.Id,
	// 	TeamID:    channel.TeamId,
	// 	Expand: &api.Expand{
	// 		Channel: api.ExpandAll,
	// 		Team:    api.ExpandAll,
	// 		User:    api.ExpandAll,
	// 	},
	// }
	// bot.dm(c.Context.ActingUserID, "Subscribed to %s in channel.", api.SubjectUserJoinedChannel)

	asBot.DM(c.Context.ActingUserID, "OK!")
	return &api.CallResponse{
		Type:     api.CallResponseTypeOK,
		Markdown: md.Markdownf("installed %s to %s channel", AppDisplayName, AppDisplayName),
	}
}
