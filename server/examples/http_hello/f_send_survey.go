package http_hello

import (
	"net/http"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

func (h *helloapp) newSendSurveyFormResponse(claims *api.JWTClaims, c *api.Call) *api.CallResponse {
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

func (h *helloapp) fSendSurvey(w http.ResponseWriter, req *http.Request, claims *api.JWTClaims, c *api.Call) (int, error) {
	var out *api.CallResponse

	switch c.Type {
	case api.CallTypeForm:
		out = h.newSendSurveyFormResponse(claims, c)

	case api.CallTypeSubmit:
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

		out = &api.CallResponse{}

		err := h.sendSurvey(userID, message)
		if err != nil {
			out.Error = err.Error()
			out.Type = api.CallResponseTypeError
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
	p.AddProp(api.PropAppBindings, []*api.Binding{
		{
			Location: "survey",
			Form:     h.newSurveyForm(message),
		},
	})
	_, err := h.dmPost(userID, p)
	return err
}
