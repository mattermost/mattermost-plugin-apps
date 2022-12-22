package proxy

import (
	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/server/config"
)

type ExpandGetter interface {
	GetChannel(channelID string) (*model.Channel, error)
	GetChannelMember(channelID, userID string) (*model.ChannelMember, error)
	GetPost(postID string) (*model.Post, error)
	GetTeam(teamID string) (*model.Team, error)
	GetTeamMember(teamID, userID string) (*model.TeamMember, error)
	GetUser(userID string) (*model.User, error)
}

type expandHTTPGetter struct {
	mm *model.Client4
}

func newExpandHTTPGetter(conf config.Config, token string) *expandHTTPGetter {
	client := model.NewAPIv4Client(conf.MattermostLocalURL)
	client.SetToken(token)
	return &expandHTTPGetter{client}
}

func (h *expandHTTPGetter) GetUser(userID string) (*model.User, error) {
	user, _, err := h.mm.GetUser(userID, "")
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (h *expandHTTPGetter) GetChannel(channelID string) (*model.Channel, error) {
	channel, _, err := h.mm.GetChannel(channelID, "")
	if err != nil {
		return nil, err
	}

	return channel, nil
}

func (h *expandHTTPGetter) GetChannelMember(channelID, userID string) (*model.ChannelMember, error) {
	channelMember, _, err := h.mm.GetChannelMember(channelID, userID, "")
	if err != nil {
		return nil, err
	}

	return channelMember, nil
}

func (h *expandHTTPGetter) GetTeam(teamID string) (*model.Team, error) {
	team, _, err := h.mm.GetTeam(teamID, "")
	if err != nil {
		return nil, err
	}

	return team, nil
}

func (h *expandHTTPGetter) GetTeamMember(teamID, userID string) (*model.TeamMember, error) {
	teamMember, _, err := h.mm.GetTeamMember(teamID, userID, "")
	if err != nil {
		return nil, err
	}

	return teamMember, nil
}

func (h *expandHTTPGetter) GetPost(postID string) (*model.Post, error) {
	post, _, err := h.mm.GetPost(postID, "")
	if err != nil {
		return nil, err
	}

	return post, nil
}

type expandRPCGetter struct {
	mm *pluginapi.Client
}

func newExpandRPCGetter(c *pluginapi.Client) *expandRPCGetter {
	return &expandRPCGetter{c}
}

func (r *expandRPCGetter) GetUser(userID string) (*model.User, error) {
	return r.mm.User.Get(userID)
}

func (r *expandRPCGetter) GetChannel(channelID string) (*model.Channel, error) {
	return r.mm.Channel.Get(channelID)
}

func (r *expandRPCGetter) GetChannelMember(channelID, userID string) (*model.ChannelMember, error) {
	return r.mm.Channel.GetMember(channelID, userID)
}

func (r *expandRPCGetter) GetTeam(teamID string) (*model.Team, error) {
	return r.mm.Team.Get(teamID)
}

func (r *expandRPCGetter) GetTeamMember(teamID, userID string) (*model.TeamMember, error) {
	return r.mm.Team.GetMember(teamID, userID)
}

func (r *expandRPCGetter) GetPost(postID string) (*model.Post, error) {
	return r.mm.Post.GetPost(postID)
}

// To work around data access timing when joining/leaving teams and channels we
// special-case the "self" events: the user, channel, team, and membership data
// are expanded bypassiing the usual permission checks. This is ok since the
// subscriber is the user in the event, and therefore can have access to the
// data.
type expandSelfGetter struct {
	ExpandGetter

	mm         *pluginapi.Client
	memberUser *model.User
	cm         *model.ChannelMember
	tm         *model.TeamMember
	channel    *model.Channel
}

func newExpandSelfGetter(mm *pluginapi.Client, memberUser *model.User, cm *model.ChannelMember, tm *model.TeamMember, channel *model.Channel) ExpandGetter {
	return &expandSelfGetter{
		ExpandGetter: newExpandRPCGetter(mm),
		mm:           mm,
		memberUser:   memberUser,
		cm:           cm,
		tm:           tm,
		channel:      channel,
	}
}

func (g *expandSelfGetter) GetUser(userID string) (*model.User, error) {
	// Bypass permission checks, since the user is self. Use the cached data if
	// available.
	if g.memberUser != nil && g.memberUser.Id == userID {
		return g.memberUser, nil
	}
	return g.mm.User.Get(userID)
}

func (g *expandSelfGetter) GetChannel(channelID string) (*model.Channel, error) {
	// Bypass permission checks, since the user is/just was in the channel. Use
	// the cached data if available.
	if g.channel != nil && g.channel.Id == channelID {
		return g.channel, nil
	}
	return g.mm.Channel.Get(channelID)
}

func (g *expandSelfGetter) GetChannelMember(channelID, userID string) (*model.ChannelMember, error) {
	// Bypass permission checks, since the user is/just was in the channel. Use
	// the cached data if available.
	if g.cm != nil && g.cm.ChannelId == channelID && g.cm.UserId == userID {
		return g.cm, nil
	}
	return g.mm.Channel.GetMember(channelID, userID)
}

func (g *expandSelfGetter) GetTeam(teamID string) (*model.Team, error) {
	// Bypass permission checks, since the user is the subscriber and is/just
	// was in the team.
	return g.mm.Team.Get(teamID)
}

func (g *expandSelfGetter) GetTeamMember(teamID, userID string) (*model.TeamMember, error) {
	// Bypass permission checks, since the user is/just was in the team. Use the
	// cached data if available.
	if g.tm != nil && g.tm.TeamId == teamID && g.tm.UserId == userID {
		return g.tm, nil
	}
	return g.mm.Team.GetMember(teamID, userID)
}
