package helloapp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
	"github.com/mattermost/mattermost-server/v5/model"
)

const (
	defaultPingMessage = "Hello!"
	callValueUserID    = "user_id"
	callValueMessage   = "message"
	callValuePostID    = "post_id"
)

func (h *helloapp) handlePing(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, call *apps.Call) (int, error) {
	userID := call.Values.Get(callValueUserID)
	if userID == "" {
		userID = call.Context.ActingUserID
	}

	fromUserID := ""
	if userID != call.Context.ActingUserID {
		fromUserID = call.Context.ActingUserID
	}

	postID := call.Values.Get(callValuePostID)
	if postID == "" && call.From[0].LocationType == apps.LocationPostMenuItem {
		postID = call.Context.PostID
	}

	finalMessage := ""
	if postID != "" {
		post, err := h.getPost(postID, claims.ActingUserID)
		if err != nil {
			return http.StatusInternalServerError, err
		}
		finalMessage = post.Message
	}

	message := call.Values.Get(callValueMessage)

	if message == "" && finalMessage == "" {
		message = defaultPingMessage
	}

	finalMessage = strings.Join([]string{message, finalMessage}, "\n")

	h.ping(userID, fromUserID, finalMessage)

	response := apps.CallResponse{
		Type: apps.CallResponseTypeOK,
		Data: make(map[string]interface{}),
	}

	if call.From[0].LocationType == apps.LocationEmbeddedForm {
		u, err := h.getUser(userID)
		toText := ""
		if err == nil {
			toText = fmt.Sprintf("\n\nto @%s", u.Username)
		}
		post := &model.Post{
			Message: fmt.Sprintf("You have sent:\n > %s%s", finalMessage, toText),
			Props:   model.StringInterface{},
		}
		response.Data["post"] = post
	}

	httputils.WriteJSON(w, response)
	return http.StatusOK, nil
}

func (h *helloapp) handleOpenPingDialog(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, call *apps.Call) (int, error) {
	postID := call.Values.Get(callValuePostID)
	if postID == "" {
		postID = call.Context.PostID
	}

	message := defaultPingMessage
	if postID != "" {
		post, err := h.getPost(postID, claims.ActingUserID)
		if err != nil {
			return http.StatusInternalServerError, err
		}
		message = post.Message
	}

	dialogID, err := h.storeDialog(h.getDialogPing(false, message))
	if err != nil {
		return http.StatusInternalServerError, err
	}

	httputils.WriteJSON(w, apps.CallResponse{
		Type: apps.CallResponseTypeCommand,
		Data: map[string]interface{}{
			"command": fmt.Sprintf("/apps openDialog %s %s %s", appID, h.appURL(pathDialogs), dialogID),
			"args": model.CommandArgs{
				ChannelId: call.Context.ChannelID,
				TeamId:    call.Context.TeamID,
			},
		},
	})
	return http.StatusOK, nil
}

func (h *helloapp) handleCreatePingEmbedded(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, call *apps.Call) (int, error) {
	postID := call.Values.Get(callValuePostID)
	if postID == "" {
		postID = call.Context.PostID
	}
	message := defaultPingMessage
	if postID != "" {
		post, err := h.getPost(postID, claims.ActingUserID)
		if err != nil {
			return http.StatusInternalServerError, err
		}
		message = post.Message
	}

	userID := claims.ActingUserID
	if call.Context.UserID != "" {
		userID = call.Context.UserID
	}

	h.dmPost(userID, &model.Post{
		Message: "Let me ping someone!",
		Props: model.StringInterface{
			"dialog": h.getDialogPing(true, message),
			"appID":  appID,
		},
	})

	return http.StatusOK, nil
}

func (h *helloapp) ping(userID string, from string, message string) {
	if from != "" {
		user, err := h.getUser(from)
		if err == nil {
			message = fmt.Sprintf("@%s pings you with: %s", user.Username, message)
		}
	}
	_, _ = h.dm(userID, message)
}

func (h *helloapp) handleSubmitPingDialog(w http.ResponseWriter, req *http.Request) {
	response := model.SubmitDialogResponse{
		Errors: make(map[string]string),
	}
	defer httputils.WriteJSON(w, response)

	actingUserID := req.Header.Get("Mattermost-User-Id")
	if actingUserID == "" {
		response.Error = "User not logged in."
		return
	}

	var dialogSubmission model.SubmitDialogRequest
	err := json.NewDecoder(req.Body).Decode(&dialogSubmission)
	if err != nil {
		response.Error = "Cannot decode submission."
		return
	}

	userID, ok := dialogSubmission.Submission[dialogFieldUserID].(string)
	if !ok {
		response.Errors[dialogFieldUserID] = "User is required."
		return
	}
	message, ok := dialogSubmission.Submission[dialogFieldMessage].(string)
	if !ok {
		response.Errors[dialogFieldMessage] = "Message required."
		return
	}
	h.ping(userID, actingUserID, message)
}
