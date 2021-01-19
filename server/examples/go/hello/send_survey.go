package hello

import (
	"fmt"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/modelapps"
	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

type SurveyFormSubmission struct {
	UserID  string                 `json:"userID"`
	Message string                 `json:"message"`
	Other   map[string]interface{} `json:"other"`
}

func extractSurveyFormValues(c *modelapps.Call) SurveyFormSubmission {
	message := ""
	userID := ""
	var other map[string]interface{} = nil
	if c.Context != nil && c.Context.Post != nil {
		message = c.Context.Post.Message
	}

	topValues := c.Values
	formValues := c.Values
	if c.Type == modelapps.CallTypeForm && topValues != nil {
		formValues, _ = topValues["values"].(map[string]interface{})
	}

	if formValues != nil {
		userID, _ = formValues["userID"].(string)
		message, _ = formValues["message"].(string)
		otherTemp, ok2 := formValues["other"].(map[string]interface{})
		if ok2 {
			other = otherTemp
		} else {
			other = nil
		}
	}

	return SurveyFormSubmission{
		UserID:  userID,
		Message: message,
		Other:   other,
	}
}

func NewSendSurveyFormResponse(c *modelapps.Call) *modelapps.CallResponse {
	submission := extractSurveyFormValues(c)
	name, _ := c.Values["name"].(string)

	if name == "userID" {
		submission.Message = fmt.Sprintf("%s Now sending to %s.", submission.Message, submission.UserID)
	}

	return &modelapps.CallResponse{
		Type: modelapps.CallResponseTypeForm,
		Form: &modelapps.Form{
			Title:  "Send a survey to user",
			Header: "Message modal form header",
			Footer: "Message modal form footer",
			Call:   modelapps.MakeCall(PathSendSurvey),
			Fields: []*modelapps.Field{
				{
					Name:                 fieldUserID,
					Type:                 modelapps.FieldTypeUser,
					Description:          "User to send the survey to",
					Label:                "user",
					ModalLabel:           "User",
					AutocompleteHint:     "enter user ID or @user",
					AutocompletePosition: 1,
					Value:                submission.UserID,
					SelectRefresh:        true,
				}, {
					Name:             "other",
					Type:             modelapps.FieldTypeDynamicSelect,
					Description:      "Some values",
					Label:            "other",
					AutocompleteHint: "Pick one",
					ModalLabel:       "Other",
					Value:            submission.Other,
				}, {
					Name:             fieldMessage,
					Type:             modelapps.FieldTypeText,
					Description:      "Text to ask the user about",
					IsRequired:       true,
					Label:            "message",
					ModalLabel:       "Text",
					AutocompleteHint: "Anything you want to say",
					TextSubtype:      "textarea",
					TextMinLength:    2,
					TextMaxLength:    1024,
					Value:            submission.Message,
				},
			},
		},
	}
}

func NewSendSurveyPartialFormResponse(c *modelapps.Call) *modelapps.CallResponse {
	if c.Type == modelapps.CallTypeSubmit {
		return NewSendSurveyFormResponse(c)
	}

	return &modelapps.CallResponse{
		Type: modelapps.CallResponseTypeForm,
		Form: &modelapps.Form{
			Title:  "Send a survey to user",
			Header: "Message modal form header",
			Footer: "Message modal form footer",
			Call:   modelapps.MakeCall(PathSendSurveyCommandToModal),
			Fields: []*modelapps.Field{
				{
					Name:             fieldMessage,
					Type:             modelapps.FieldTypeText,
					Description:      "Text to ask the user about",
					IsRequired:       true,
					Label:            "message",
					ModalLabel:       "Text",
					AutocompleteHint: "Anything you want to say",
					TextSubtype:      "textarea",
					TextMinLength:    2,
					TextMaxLength:    1024,
					Value:            "",
				},
			},
		},
	}
}

func (h *HelloApp) SendSurvey(c *modelapps.Call) (md.MD, error) {
	bot := modelapps.AsBot(c.Context)
	userID := c.GetValue(fieldUserID, c.Context.ActingUserID)

	// TODO this should be done with expanding mentions, make a ticket
	if strings.HasPrefix(userID, "@") {
		user, _ := bot.GetUserByUsername(userID[1:], "")
		if user != nil {
			userID = user.Id
		}
	}

	message := c.GetValue(fieldMessage, "Hello")
	if c.Context.Post != nil {
		message += "\n>>> " + c.Context.Post.Message
	}

	err := sendSurvey(bot, userID, message)
	if err != nil {
		return "", err
	}

	return "Successfully sent survey", nil
}

func sendSurvey(bot modelapps.Client, userID, message string) error {
	p := &model.Post{
		Message: "Please respond to this survey: " + message,
	}
	p.AddProp(api.PropAppBindings, []*modelapps.Binding{
		{
			Location: "survey",
			Form:     NewSurveyForm(message),
		},
	})
	_, err := bot.DMPost(userID, p)
	return err
}
