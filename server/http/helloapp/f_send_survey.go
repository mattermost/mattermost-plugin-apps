package helloapp

import (
	"net/http"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

func (h *helloapp) newSendSurveyFormResponse(claims *apps.JWTClaims, c *apps.Call) *apps.CallResponse {
	message := ""
	if c.Context != nil && c.Context.Post != nil {
		message = c.Context.Post.Message
	}

	return &apps.CallResponse{
		Type: apps.CallResponseTypeForm,
		Form: &apps.Form{
			Title:  "Send a survey to user",
			Header: "Message modal form header",
			// Footer: "Message modal form footer",
			Fields: []*apps.Field{
				{
					Name:             fieldUserID,
					Type:             apps.FieldTypeUser,
					Description:      "User to send the survey to",
					Label:            "user",
					AutocompleteHint: "enter user ID or @user",
					ModalLabel:       "User",
				}, {
					Name:             fieldMessage,
					Type:             apps.FieldTypeText,
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

func (h *helloapp) fSendSurvey(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, c *apps.Call) (int, error) {
	var out *apps.CallResponse

	switch c.Type {
	case apps.CallTypeForm:
		out = h.newSendSurveyFormResponse(claims, c)

	case apps.CallTypeSubmit:
		userID := c.GetValue(fieldUserID, c.Context.ActingUserID)

		// TODO this should be done with expanding mentions, make a ticket
		if strings.HasPrefix(userID, "@") {
			_ = h.asUser(c.Context.ActingUserID, func(c *model.Client4) error {
				user, _ := c.GetUserByUsername(userID[1:], "")
				if user != nil {
					userID = user.Id
				}
				return nil
			})
		}

		message := c.GetValue(fieldMessage, "Hello")
		if c.Context.Post != nil {
			message += "\n>>> " + c.Context.Post.Message
		}

		out = &apps.CallResponse{}

		err := h.sendSurvey(userID, message)
		if err != nil {
			out.Error = err.Error()
			out.Type = apps.CallResponseTypeError
		} else {
			out.Markdown = md.Markdownf(
				"Successfully sent survey",
			)
		}
	}
	httputils.WriteJSON(w, out)
	return http.StatusOK, nil
}

func (h *helloapp) sendSurvey(userID, message string) error {
	p := &model.Post{
		Message: "Please respond to this survey: " + message,
	}
	p.AddProp(apps.PropAppBindings, []*apps.Binding{
		{
			Location: "survey",
			Form:     h.newSurveyForm(message),
		},
	})
	_, err := h.dmPost(userID, p)
	return err
}
