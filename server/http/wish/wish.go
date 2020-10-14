package wish

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/constants"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

type wish struct {
	apps *apps.Service
}

func Init(router *mux.Router, apps *apps.Service) {
	w := wish{
		apps: apps,
	}

	subrouter := router.PathPrefix(constants.WishPath).Subrouter()
	subrouter.HandleFunc("", w.handleWish).Methods("POST")
}

func (wish *wish) verifyCallContext(actingUserID string, ctx apps.CallContext) error {
	if ctx.ChannelID == "" {
		return errors.New("no channel id provided in call")
	}

	_, err := wish.apps.Mattermost.Channel.GetMember(ctx.ChannelID, actingUserID)
	if err != nil {
		err = errors.Errorf("user is not a member of the specified channel. user=%v channel=%v", actingUserID, ctx.ChannelID)
		return err
	}

	return nil
}

func (wish *wish) handleWish(w http.ResponseWriter, req *http.Request) {
	var err error

	call := apps.Call{}
	defer req.Body.Close()

	err = json.NewDecoder(req.Body).Decode(&call)
	if err != nil {
		err = errors.Wrap(err, "Failed to unmarshal Call struct")
		httputils.WriteBadRequestError(w, err)
		return
	}

	ctx := call.Data.Context
	if ctx.AppID == "" {
		err = errors.New("no app id provided in context from frontend")
		httputils.WriteBadRequestError(w, err)
	}

	actingUserID := req.Header.Get("Mattermost-User-Id")
	if actingUserID == "" {
		err = errors.New("user not logged in")
		httputils.WriteUnauthorizedError(w, err)
		return
	}

	ctx.ActingUserID = actingUserID

	err = wish.verifyCallContext(actingUserID, ctx)
	if err != nil {
		err = errors.Wrap(err, "failed to verify call context")
		httputils.WriteBadRequestError(w, err)
		return
	}

	// TODO: Support predefined expand levels per wish. We don't want the client to decide what to expand, or do we?
	expanded, err := wish.apps.Expander.Expand(&apps.Expand{
		User: apps.ExpandAll,
	}, actingUserID, actingUserID, call.Data.Context.ChannelID)
	if err != nil {
		httputils.WriteJSONError(w, http.StatusInternalServerError, "Failed to expand", err)
		return
	}

	call.Data.Expanded = expanded

	res, err := wish.apps.Call(call)
	if err != nil {
		err = errors.Wrap(err, "Call to external integration server failed")
		httputils.WriteInternalServerError(w, err)
		return
	}

	httputils.WriteJSON(w, res)
}
