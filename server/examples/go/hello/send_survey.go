package hello

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/mmclient"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

type SurveyFormSubmission struct {
	UserID  string                 `json:"userID"`
	Message string                 `json:"message"`
	Other   map[string]interface{} `json:"other"`
}

func mapToSurveyFormSubmission(values map[string]interface{}) SurveyFormSubmission {
	submission := SurveyFormSubmission{}
	b, err := json.Marshal(values)
	if err != nil {
		return submission
	}

	err = json.Unmarshal(b, &submission)
	if err != nil {
		return submission
	}

	return submission
}

func extractSurveyFormValues(c *apps.CallRequest) SurveyFormSubmission {
	submission := mapToSurveyFormSubmission(c.Values)

	if submission.Message == "" && c.Context != nil && c.Context.Post != nil {
		submission.Message = c.Context.Post.Message
	}

	return submission
}

func NewSendSurveyFormResponse(c *apps.CallRequest) *apps.CallResponse {
	submission := extractSurveyFormValues(c)
	name := c.SelectedField

	if name == "userID" {
		submission.Message = fmt.Sprintf("%s Now sending to %s.", submission.Message, submission.UserID)
	}

	return &apps.CallResponse{
		Type: apps.CallResponseTypeForm,
		Form: &apps.Form{
			Title:  "Send a survey to user",
			Header: "Message modal form header",
			Footer: "Message modal form footer",
			Call:   apps.MakeCall(PathSendSurvey),
			Fields: []*apps.Field{
				{
					Name:                 fieldUserID,
					Type:                 apps.FieldTypeUser,
					Description:          "User to send the survey to",
					Label:                "user",
					ModalLabel:           "User",
					AutocompleteHint:     "enter user ID or @user",
					AutocompletePosition: 1,
					Value:                submission.UserID,
					SelectRefresh:        true,
				}, {
					Name:             "other",
					Type:             apps.FieldTypeDynamicSelect,
					Description:      "Some values",
					Label:            "other",
					AutocompleteHint: "Pick one",
					ModalLabel:       "Other",
					Value:            submission.Other,
				}, {
					Name:             fieldMessage,
					Type:             apps.FieldTypeText,
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

func NewSendSurveyPartialFormResponse(c *apps.CallRequest) *apps.CallResponse {
	if c.Type == apps.CallTypeSubmit {
		return NewSendSurveyFormResponse(c)
	}

	return &apps.CallResponse{
		Type: apps.CallResponseTypeForm,
		Form: &apps.Form{
			Title:  "Send a survey to user",
			Header: "Message modal form header",
			Footer: "Message modal form footer",
			Call:   apps.MakeCall(PathSendSurveyCommandToModal),
			Fields: []*apps.Field{
				{
					Name:             fieldMessage,
					Type:             apps.FieldTypeText,
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

func (h *HelloApp) SendSurvey(c *apps.CallRequest) (md.MD, error) {
	bot := mmclient.AsBot(c.Context)
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

func sendSurvey(bot *mmclient.Client, userID, message string) error {
	p := &model.Post{
		Message: "Please respond to this survey: " + message,
	}
	p.AddProp(apps.PropAppBindings, []*apps.Binding{
		{
			AppID:       "http-hello",
			Location:    "survey",
			Label:       "Survey",
			Description: message,
			Bindings: []*apps.Binding{
				{
					Location: "select",
					Label:    "Select one",
					Call:     apps.MakeCall(PathSubmitSurvey),
					Bindings: []*apps.Binding{
						{
							Location: "good",
							Label:    "Good",
						},
						{
							Location: "normal",
							Label:    "Normal",
						},
						{
							Location: "bad",
							Label:    "Bad",
						},
					},
				},
				{
					Location: "button",
					Label:    "Do not send",
					Call:     apps.MakeCall(PathSubmitSurvey),
				},
			},
		},
	})
	_, err := bot.DMPost(userID, p)
	return err
}
