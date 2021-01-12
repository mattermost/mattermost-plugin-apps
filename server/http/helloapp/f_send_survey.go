package helloapp

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

type SurveyFormSubmission struct {
	UserID  string                 `json:"userID"`
	Message string                 `json:"message"`
	Other   map[string]interface{} `json:"other"`
}

func extractSurveyFormValues(c *apps.Call) SurveyFormSubmission {
	message := ""
	userID := ""
	var other map[string]interface{} = nil
	if c.Context != nil && c.Context.Post != nil {
		message = c.Context.Post.Message
	}

	topValues := c.Values
	if topValues != nil && topValues["values"] != nil {
		if values, ok := topValues["values"].(map[string]interface{}); ok {
			userID, _ = values["userID"].(string)
			message, _ = values["message"].(string)
			otherTemp, ok2 := values["other"].(map[string]interface{})
			if ok2 {
				other = otherTemp
			} else {
				other = nil
			}
		}
	}

	return SurveyFormSubmission{
		UserID:  userID,
		Message: message,
		Other:   other,
	}
}

func (h *helloapp) newSendSurveyFormResponse(claims *apps.JWTClaims, c *apps.Call) *apps.CallResponse {
	submission := extractSurveyFormValues(c)
	name, _ := c.Values["name"].(string)

	if name == "userID" {
		submission.Message = fmt.Sprintf("%s Now sending to %s.", submission.Message, submission.UserID)
	}

	return &apps.CallResponse{
		Type: apps.CallResponseTypeForm,
		Form: &apps.Form{
			Title:  "Send a survey to user",
			Header: "Message modal form header",
			Footer: "Message modal form footer",
			Fields: []*apps.Field{
				{
					Name:             fieldUserID,
					Type:             apps.FieldTypeUser,
					Description:      "User to send the survey to",
					Label:            "user",
					AutocompleteHint: "enter user ID or @user",
					ModalLabel:       "User",
					SelectRefresh:    true,
					Value:            submission.UserID,
				}, {
					Name:             "other",
					Type:             apps.FieldTypeDynamicSelect,
					Description:      "Some values",
					Label:            "other",
					AutocompleteHint: "Pick one",
					ModalLabel:       "Other",
					SelectRefresh:    true,
					Value:            submission.Other,
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
					Value:            submission.Message,
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
	case apps.CallTypeLookup:
		out = &apps.CallResponse{
			Data: map[string]interface{}{
				"items": []*apps.SelectOption{
					{
						Label: "Option 1",
						Value: "option1",
					},
				},
			},
		}

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
