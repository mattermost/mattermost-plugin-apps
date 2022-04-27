package restapi

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/httpin"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

var emptyCC = apps.Context{}

func (a *restapi) initCall(h *httpin.Handler) {
	h.HandleFunc(path.Call,
		a.Call, httpin.RequireUser).Methods(http.MethodPost)
}

// Call handles a call request for an App.
//   Path: /api/v1/call
//   Method: POST
//   Input: CallRequest
//   Output: CallResponse
func (a *restapi) Call(r *incoming.Request, w http.ResponseWriter, req *http.Request) {
	creq, err := apps.CallRequestFromJSONReader(req.Body)
	if err != nil {
		httputils.WriteError(w, utils.NewInvalidError(errors.Wrap(err, "failed to unmarshal Call request")))
		return
	}
	r.SetAppID(creq.Context.AppID)

	// Clear out anythging in the incoming expanded context for security
	// reasons, it will be set by Expand before passing to the app.
	creq.Context.ExpandedContext = apps.ExpandedContext{}
	creq.Context, err = a.cleanUserAgentContext(r.ActingUserID(), creq.Context)
	if err != nil {
		httputils.WriteError(w, utils.NewInvalidError(errors.Wrap(err, "invalid call context for user")))
		return
	}

	cresp := a.proxy.Call(r, *creq)

	// Add the request and response digests to the logger.
	r.Log = r.Log.With(creq, cresp)

	// Only track submit calls.
	if creq.Context.UserAgentContext.TrackAsSubmit {
		a.conf.Telemetry().TrackCall(string(creq.Context.AppID), string(creq.Context.Location), creq.Context.ActingUserID, "submit")
	}

	_ = httputils.WriteJSON(w, cresp)
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
