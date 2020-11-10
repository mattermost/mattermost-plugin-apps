package helloapp

import (
	"net/http"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

func (h *helloapp) newSendSurveyFormResponse(claims *apps.JWTClaims, c *apps.Call) *apps.CallResponse {
	return &apps.CallResponse{
		Type: apps.CallResponseTypeForm,
		Form: &apps.Form{
			Title:  "Send a survey to user",
			Header: "Message modal form header",
			Footer: "Message modal form footer",
			Fields: []*apps.Field{
				{
					Name:                 fieldUserID,
					Type:                 apps.FieldTypeUser,
					Description:          "User to send the survey to",
					Label:                "User",
					AutocompleteHint:     "enter user ID or @user",
					AutocompletePosition: 1,
					ModalLabel:           "User",
				}, {
					Name:             fieldMessage,
					Type:             apps.FieldTypeText,
					IsRequired:       true,
					Description:      "Text to ask the user about",
					Label:            "message",
					AutocompleteHint: "Anything you want to say",
					ModalLabel:       "Text",
					TextMinLength:    2,
					TextMaxLength:    1024,
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

		h.sendSurvey(userID, message)
		out = &apps.CallResponse{}
	}
	httputils.WriteJSON(w, out)
	return http.StatusOK, nil
}

func (h *helloapp) sendSurvey(userID, message string) {
	p := &model.Post{
		Message: "Please respond to this survey",
	}
	p.AddProp(apps.PropAppBindings, []*apps.Binding{
		{
			LocationID: "survey",
			Form:       h.newSurveyForm(message),
		},
	})
	h.dmPost(userID, p)
}
