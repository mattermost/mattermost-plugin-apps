package proxy

import (
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin/pluginapi"
	"golang.org/x/net/context"

	"github.com/mattermost/mattermost-plugin-apps/server/config"
)

type ExpandGetter interface {
	GetChannel(ctx context.Context, channelID string) (*model.Channel, error)
	GetChannelMember(ctx context.Context, channelID, userID string) (*model.ChannelMember, error)
	GetPost(ctx context.Context, postID string) (*model.Post, error)
	GetTeam(ctx context.Context, teamID string) (*model.Team, error)
	GetTeamMember(ctx context.Context, teamID, userID string) (*model.TeamMember, error)
	GetUser(ctx context.Context, userID string) (*model.User, error)
}

type expandHTTPGetter struct {
	mm *model.Client4
}

func newExpandHTTPGetter(conf config.Config, token string) *expandHTTPGetter {
	client := model.NewAPIv4Client(conf.MattermostLocalURL)
	client.SetToken(token)
	return &expandHTTPGetter{client}
}

func (h *expandHTTPGetter) GetUser(ctx context.Context, userID string) (*model.User, error) {
	user, _, err := h.mm.GetUser(ctx, userID, "")
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (h *expandHTTPGetter) GetChannel(ctx context.Context, channelID string) (*model.Channel, error) {
	channel, _, err := h.mm.GetChannel(ctx, channelID, "")
	if err != nil {
		return nil, err
	}

	return channel, nil
}

func (h *expandHTTPGetter) GetChannelMember(ctx context.Context, channelID, userID string) (*model.ChannelMember, error) {
	channelMember, _, err := h.mm.GetChannelMember(ctx, channelID, userID, "")
	if err != nil {
		return nil, err
	}

	return channelMember, nil
}

func (h *expandHTTPGetter) GetTeam(ctx context.Context, teamID string) (*model.Team, error) {
	team, _, err := h.mm.GetTeam(ctx, teamID, "")
	if err != nil {
		return nil, err
	}

	return team, nil
}

func (h *expandHTTPGetter) GetTeamMember(ctx context.Context, teamID, userID string) (*model.TeamMember, error) {
	teamMember, _, err := h.mm.GetTeamMember(ctx, teamID, userID, "")
	if err != nil {
		return nil, err
	}

	return teamMember, nil
}

func (h *expandHTTPGetter) GetPost(ctx context.Context, postID string) (*model.Post, error) {
	post, _, err := h.mm.GetPost(ctx, postID, "")
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

func (r *expandRPCGetter) GetUser(_ context.Context, userID string) (*model.User, error) {
	return r.mm.User.Get(userID)
}

func (r *expandRPCGetter) GetChannel(_ context.Context, channelID string) (*model.Channel, error) {
	return r.mm.Channel.Get(channelID)
}

func (r *expandRPCGetter) GetChannelMember(_ context.Context, channelID, userID string) (*model.ChannelMember, error) {
	return r.mm.Channel.GetMember(channelID, userID)
}

func (r *expandRPCGetter) GetTeam(_ context.Context, teamID string) (*model.Team, error) {
	return r.mm.Team.Get(teamID)
}

func (r *expandRPCGetter) GetTeamMember(_ context.Context, teamID, userID string) (*model.TeamMember, error) {
	return r.mm.Team.GetMember(teamID, userID)
}

func (r *expandRPCGetter) GetPost(_ context.Context, postID string) (*model.Post, error) {
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

func (g *expandSelfGetter) GetUser(_ context.Context, userID string) (*model.User, error) {
	// Bypass permission checks, since the user is self. Use the cached data if
	// available.
	if g.memberUser != nil && g.memberUser.Id == userID {
		return g.memberUser, nil
	}
	return g.mm.User.Get(userID)
}

func (g *expandSelfGetter) GetChannel(_ context.Context, channelID string) (*model.Channel, error) {
	// Bypass permission checks, since the user is/just was in the channel. Use
	// the cached data if available.
	if g.channel != nil && g.channel.Id == channelID {
		return g.channel, nil
	}
	return g.mm.Channel.Get(channelID)
}

func (g *expandSelfGetter) GetChannelMember(_ context.Context, channelID, userID string) (*model.ChannelMember, error) {
	// Bypass permission checks, since the user is/just was in the channel. Use
	// the cached data if available.
	if g.cm != nil && g.cm.ChannelId == channelID && g.cm.UserId == userID {
		return g.cm, nil
	}
	return g.mm.Channel.GetMember(channelID, userID)
}

func (g *expandSelfGetter) GetTeam(_ context.Context, teamID string) (*model.Team, error) {
	// Bypass permission checks, since the user is the subscriber and is/just
	// was in the team.
	return g.mm.Team.Get(teamID)
}

func (g *expandSelfGetter) GetTeamMember(_ context.Context, teamID, userID string) (*model.TeamMember, error) {
	// Bypass permission checks, since the user is/just was in the team. Use the
	// cached data if available.
	if g.tm != nil && g.tm.TeamId == teamID && g.tm.UserId == userID {
		return g.tm, nil
	}
	return g.mm.Team.GetMember(teamID, userID)
}
