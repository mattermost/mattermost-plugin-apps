package hello

import (
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/examples"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

func NewSendSurveyFormResponse(c *api.Call) *api.CallResponse {
	message := ""
	if c.Context != nil && c.Context.Post != nil {
		message = c.Context.Post.Message
	}

	return &api.CallResponse{
		Type: api.CallResponseTypeForm,
		Form: &api.Form{
			Title:  "Send a survey to user",
			Header: "Message modal form header",
			Footer: "Message modal form footer",
			Fields: []*api.Field{
				{
					Name:                 fieldUserID,
					Type:                 api.FieldTypeUser,
					Description:          "User to send the survey to",
					Label:                "user",
					ModalLabel:           "User",
					AutocompleteHint:     "enter user ID or @user",
					AutocompletePosition: 1,
				}, {
					Name:             fieldMessage,
					Type:             api.FieldTypeText,
					Description:      "Text to ask the user about",
					IsRequired:       true,
					Label:            "message",
					ModalLabel:       "Text",
					AutocompleteHint: "Anything you want to say",
					TextSubtype:      "textarea",
					TextMinLength:    2,
					TextMaxLength:    1024,
					Value:            message,
				},
			},
		},
	}
}

func (h *HelloApp) SendSurvey(c *api.Call) (md.MD, error) {
	bot := examples.AsBot(c.Context)
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

func sendSurvey(bot examples.Client, userID, message string) error {
	p := &model.Post{
		Message: "Please respond to this survey: " + message,
	}
	p.AddProp(api.PropAppBindings, []*api.Binding{
		{
			Location: "survey",
			Form:     NewSurveyForm(message),
		},
	})
	_, err := bot.DMPost(userID, p)
	return err
}
