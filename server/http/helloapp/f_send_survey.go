package helloapp

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/constants"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

func (h *helloapp) SendSurvey(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, c *api.Call) (int, error) {
	var out *api.CallResponse

	switch c.Type {
	// TODO: Not yet used, will be for commands, Modals
	case api.CallTypeForm:
		out = &api.CallResponse{
			Type: api.CallResponseTypeForm,
			Form: &api.Form{
				Title:  "Send a survey to user",
				Header: "Message modal form header",
				Footer: "Message modal form footer",
				Fields: []*api.Field{
					{
						Name:              fieldUserID,
						Type:              api.FieldTypeUser,
						Description:       "User to send the survey to",
						AutocompleteLabel: "user",
						AutocompleteHint:  "enter user ID or @user",
						ModalLabel:        "User",
					}, {
						Name:              fieldMessage,
						Type:              api.FieldTypeText,
						IsRequired:        true,
						Description:       "Text to ask the user about",
						AutocompleteLabel: "$1",
						AutocompleteHint:  "Anything you want to say",
						ModalLabel:        "Text",
						TextMinLength:     2,
						TextMaxLength:     1024,
					},
				},
			},
		}

	case api.CallTypeSubmit:
		userID := c.GetValue(fieldUserID, c.Context.ActingUserID)
		message := c.GetValue(fieldMessage, "Hello")
		if c.Context.Post != nil {
			message += "\n>>> " + c.Context.Post.Message
		}

		h.sendSurvey(userID, message)
		out = &api.CallResponse{}
	}
	httputils.WriteJSON(w, out)
	return http.StatusOK, nil
}

func (h *helloapp) sendSurvey(userID, message string) {
	p := &model.Post{
		Message: "Please respond to this survey",
	}
	p.AddProp(constants.PostPropAppID, appID)
	p.AddProp(constants.PostPropDebugDialog, h.newSurveyDebugDialog(message))

	h.dmPost(userID, p)
}

func (h *helloapp) newSendSurveyDebugDialog(defaultMessage string) *model.OpenDialogRequest {
	url := pathSendSurvey

	return &model.OpenDialogRequest{
		TriggerId: appID,
		URL:       h.appURL(url),
		Dialog: model.Dialog{
			IconURL: "http://www.mattermost.org/wp-content/uploads/2016/04/icon.png",
			Elements: []model.DialogElement{
				{
					DisplayName: "Who do you want to ping?",
					Name:        fieldUserID,
					Type:        "select",
					DataSource:  "users",
				},
				{
					DisplayName: "What do you want to say?",
					Name:        fieldMessage,
					Type:        "text",
					Default:     defaultMessage,
				},
			},
		},
	}
}

func (h *helloapp) handleSendSurveyDebugDialogSubmit(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims) (int, error) {
	response := model.SubmitDialogResponse{
		Errors: make(map[string]string),
	}
	defer httputils.WriteJSON(w, response)

	var dialogSubmission model.SubmitDialogRequest
	err := json.NewDecoder(req.Body).Decode(&dialogSubmission)
	if err != nil {
		response.Error = "Cannot decode submission."
		return http.StatusOK, nil
	}

	userID, ok := dialogSubmission.Submission[fieldUserID].(string)
	if !ok {
		response.Errors[fieldUserID] = "User is required."
		return http.StatusOK, nil
	}
	message, ok := dialogSubmission.Submission[fieldMessage].(string)
	if !ok {
		response.Errors[fieldMessage] = "Message required."
		return http.StatusOK, nil
	}

	h.sendSurvey(userID, message)
	return http.StatusOK, nil
}

func (h *helloapp) OpenSendSurveyDebugDialog(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, c *api.Call) (int, error) {
	message := c.GetValue(fieldMessage, "Hello")
	if c.Context.Post != nil {
		message += "\n>>> " + c.Context.Post.Message
	}

	dialogID, err := h.storeDialog(h.newSendSurveyDebugDialog(message))
	if err != nil {
		return http.StatusInternalServerError, err
	}

	httputils.WriteJSON(w, api.CallResponse{
		Type: api.CallResponseTypeCommand,
		Data: map[string]interface{}{
			"command": fmt.Sprintf("/apps openDialog %s %s %s", appID, h.appURL(pathSendSurvey), dialogID),
			"args": model.CommandArgs{
				ChannelId: c.Context.ChannelID,
				TeamId:    c.Context.TeamID,
			},
		},
	})
	return http.StatusOK, nil
}
