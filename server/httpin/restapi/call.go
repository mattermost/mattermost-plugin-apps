package restapi

import (
	"net/http"
	"strings"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

func (a *restapi) handleCall(w http.ResponseWriter, req *http.Request, sessionID, actingUserID string) {
	conf, _, log := a.conf.Basic()

	call, err := apps.CallRequestFromJSONReader(req.Body)
	if err != nil {
		httputils.WriteError(w, utils.NewInvalidError(errors.Wrap(err, "failed to unmarshal Call request")))
		return
	}

	cc := call.Context

	err = a.cleanUserAgentContext(actingUserID, call.Context)
	if err != nil {
		httputils.WriteError(w, utils.NewInvalidError(errors.Wrap(err, "invalid call context for user")))
		return
	}

	cc = conf.SetContextDefaults(cc)

	call.Context = cc
	res := a.proxy.Call(sessionID, actingUserID, call)

	log.Debugw(
		"Received call response",
		"app_id", call.Context.AppID,
		"acting_user_id", call.Context.ActingUserID,
		"error", res.ErrorText,
		"type", res.Type,
		"path", call.Path,
	)

	// Only track submit calls
	if strings.HasSuffix(call.Path, "submit") {
		a.conf.Telemetry().TrackCall(string(call.Context.AppID), string(call.Context.Location), call.Context.ActingUserID, "submit")
	}

	httputils.WriteJSON(w, res)
}

func (a *restapi) cleanUserAgentContext(userID string, cc *apps.Context) error {
	mm := a.conf.MattermostAPI()
	*cc = apps.Context{
		UserAgentContext: cc.UserAgentContext,
	}

	var postID, channelID, teamID string

	switch {
	case cc.PostID != "":
		postID = cc.PostID

		post, err := mm.Post.GetPost(postID)
		if err != nil {
			return errors.Wrapf(err, "failed to get post. post=%v", postID)
		}

		channelID = post.ChannelId

		_, err = mm.Channel.GetMember(channelID, userID)
		if err != nil {
			return errors.Wrapf(err, "failed to get channel membership. user=%v channel=%v", userID, channelID)
		}

		c, err := mm.Channel.Get(channelID)
		if err != nil {
			return errors.Wrapf(err, "failed to get channel. channel=%v", channelID)
		}

		teamID = c.TeamId

	case cc.ChannelID != "":
		channelID = cc.ChannelID

		_, err := mm.Channel.GetMember(channelID, userID)
		if err != nil {
			return errors.Wrapf(err, "failed to get channel membership. user=%v channel=%v", userID, channelID)
		}

		c, err := mm.Channel.Get(channelID)
		if err != nil {
			return errors.Wrapf(err, "failed to get channel. channel=%v", channelID)
		}

		teamID = c.TeamId

	case cc.TeamID != "":
		teamID = cc.TeamID

		_, err := mm.Team.GetMember(teamID, userID)
		if err != nil {
			return errors.Wrapf(err, "failed to get team membership. user=%v team=%v", userID, teamID)
		}

	default:
		return errors.Errorf("no post, channel, or team context provided. user=%v", userID)
	}

	cc.PostID = postID
	cc.ChannelID = channelID
	cc.TeamID = teamID
	cc.ActingUserID = userID

	return nil
}
