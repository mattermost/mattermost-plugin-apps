package restapi

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

var emptyCC = apps.Context{}

func (a *restapi) handleCall(w http.ResponseWriter, req *http.Request, in proxy.Incoming) {
	creq, err := apps.CallRequestFromJSONReader(req.Body)
	if err != nil {
		httputils.WriteError(w, utils.NewInvalidError(errors.Wrap(err, "failed to unmarshal Call request")))
		return
	}

	// Clear out anythging in the incoming expanded context for security
	// reasons, it will be set by Expand before passing to the app.
	creq.Context.ExpandedContext = apps.ExpandedContext{}
	creq.Context, err = a.cleanUserAgentContext(in.ActingUserID, creq.Context)
	if err != nil {
		httputils.WriteError(w, utils.NewInvalidError(errors.Wrap(err, "invalid call context for user")))
		return
	}

	res := a.proxy.Call(in, *creq)

	a.conf.Logger().Debugw(
		"Received call response",
		"app_id", creq.Context.AppID,
		"acting_user_id", in.ActingUserID,
		"error", res.ErrorText,
		"type", res.Type,
		"path", creq.Path,
	)
	httputils.WriteJSON(w, res)
}

func (a *restapi) cleanUserAgentContext(userID string, orig apps.Context) (apps.Context, error) {
	mm := a.conf.MattermostAPI()
	var postID, channelID, teamID string
	cc := apps.Context{
		UserAgentContext: orig.UserAgentContext,
	}

	switch {
	case cc.PostID != "":
		postID = cc.PostID
		post, err := mm.Post.GetPost(postID)
		if err != nil {
			return emptyCC, errors.Wrapf(err, "failed to get post. post=%v", postID)
		}

		channelID = post.ChannelId
		_, err = mm.Channel.GetMember(channelID, userID)
		if err != nil {
			return emptyCC, errors.Wrapf(err, "failed to get channel membership. user=%v channel=%v", userID, channelID)
		}

		c, err := mm.Channel.Get(channelID)
		if err != nil {
			return emptyCC, errors.Wrapf(err, "failed to get channel. channel=%v", channelID)
		}

		teamID = c.TeamId

	case cc.ChannelID != "":
		channelID = cc.ChannelID

		_, err := mm.Channel.GetMember(channelID, userID)
		if err != nil {
			return emptyCC, errors.Wrapf(err, "failed to get channel membership. user=%v channel=%v", userID, channelID)
		}

		c, err := mm.Channel.Get(channelID)
		if err != nil {
			return emptyCC, errors.Wrapf(err, "failed to get channel. channel=%v", channelID)
		}

		teamID = c.TeamId

	case cc.TeamID != "":
		teamID = cc.TeamID

		_, err := mm.Team.GetMember(teamID, userID)
		if err != nil {
			return emptyCC, errors.Wrapf(err, "failed to get team membership. user=%v team=%v", userID, teamID)
		}

	default:
		return emptyCC, errors.Errorf("no post, channel, or team context provided. user=%v", userID)
	}

	cc.PostID = postID
	cc.ChannelID = channelID
	cc.TeamID = teamID
	cc.ActingUserID = userID
	return cc, nil
}
