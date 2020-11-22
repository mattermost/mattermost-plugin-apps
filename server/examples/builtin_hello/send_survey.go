package builtin_hello

import (
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/examples"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

func SendSurvey(c *api.Call) *api.CallResponse {
	var out *api.CallResponse

	switch c.Type {
	case api.CallTypeForm:
		out = newSendSurveyFormResponse(c)

	case api.CallTypeSubmit:
		userID := c.GetValue(fieldUserID, c.Context.ActingUserID)
		asBot := examples.AsBot(c.Context)

		// TODO this should be done with expanding mentions, make a ticket
		if strings.HasPrefix(userID, "@") {
			user, _ := asBot.GetUserByUsername(userID[1:], "")
			if user != nil {
				userID = user.Id
			}
		}

		message := c.GetValue(fieldMessage, "Hello")
		if c.Context.Post != nil {
			message += "\n>>> " + c.Context.Post.Message
		}

		out = &api.CallResponse{}
		err := sendSurvey(asBot, userID, message)
		if err != nil {
			out.Error = err.Error()
			out.Type = api.CallResponseTypeError
		} else {
			out.Markdown = md.Markdownf(
				"Successfully sent survey",
			)
		}
	}
	return out
}

func newSendSurveyFormResponse(c *api.Call) *api.CallResponse {
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
					Label:                "User",
					AutocompleteHint:     "enter user ID or @user",
					AutocompletePosition: 1,
					ModalLabel:           "User",
				}, {
					Name:             fieldMessage,
					Type:             api.FieldTypeText,
					TextSubtype:      "textarea",
					IsRequired:       true,
					Description:      "Text to ask the user about",
					Label:            "message",
					AutocompleteHint: "Anything you want to say",
					ModalLabel:       "Text",
					TextMinLength:    2,
					TextMaxLength:    1024,
					Value:            message,
				},
			},
		},
	}
}

func sendSurvey(as examples.Client, userID, message string) error {
	p := &model.Post{
		Message: "Please respond to this survey: " + message,
	}
	p.AddProp(api.PropAppBindings, []*api.Binding{
		{
			Location: "survey",
			Form:     newSurveyForm(message),
		},
	})
	_, err := as.DMPost(userID, p)
	return err
}
