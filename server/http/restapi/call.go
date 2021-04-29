package restapi

import (
	"net/http"

	"github.com/pkg/errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

func (a *restapi) handleCall(w http.ResponseWriter, req *http.Request, sessionID, actingUserID string) {
	call, err := apps.CallRequestFromJSONReader(req.Body)
	if err != nil {
		httputils.WriteError(w, utils.NewInvalidError(errors.Wrap(err, "failed to unmarshal Call request")))
		return
	}

	cc, err := cleanUserCallContext(a.mm, actingUserID, call.Context)
	if err != nil {
		httputils.WriteError(w, utils.NewInvalidError(errors.Wrap(err, "invalid call context for user")))
		return
	}

	cc = a.conf.GetConfig().SetContextDefaults(cc)

	call.Context = cc
	res := a.proxy.Call(sessionID, actingUserID, call)
	httputils.WriteJSON(w, res)
}

func cleanUserCallContext(mm *pluginapi.Client, userID string, cc *apps.Context) (*apps.Context, error) {
	cc = &apps.Context{
		ContextFromUserAgent: cc.ContextFromUserAgent,
	}

	var postID, channelID, teamID string

	switch {
	case cc.PostID != "":
		postID = cc.PostID

		post, err := mm.Post.GetPost(postID)
		if err != nil {
			return nil, err
		}

		channelID = post.ChannelId

		_, err = mm.Channel.GetMember(channelID, userID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get channel membership. user=%v channel=%v", userID, channelID)
		}

		c, err := mm.Channel.Get(channelID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get channel. channel=%v", channelID)
		}

		teamID = c.TeamId

	case cc.ChannelID != "":
		channelID = cc.ChannelID

		_, err := mm.Channel.GetMember(cc.ChannelID, userID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get channel membership. user=%v channel=%v", userID, cc.ChannelID)
		}

		c, err := mm.Channel.Get(channelID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get channel. channel=%v", channelID)
		}

		teamID = c.TeamId

	case cc.TeamID != "":
		teamID = cc.TeamID

		_, err := mm.Team.GetMember(teamID, userID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get team membership. user=%v team=%v", userID, teamID)
		}

	default:
		return nil, errors.Errorf("no post, channel, or team context provided. user=%v", userID)
	}

	cc.PostID = postID
	cc.ChannelID = channelID
	cc.TeamID = teamID

	return cc, nil
}
