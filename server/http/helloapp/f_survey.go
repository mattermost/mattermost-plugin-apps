package helloapp

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

func (h *helloapp) newSurveyForm(message string) *apps.Form {
	return &apps.Form{
		Title:  "Emotional response survey",
		Header: message,
		Fields: []*apps.Field{
			{
				Label: "What is your vote for today?",
				Name:  "vote",
				Type:  apps.FieldTypeButton,
				SelectStaticOptions: []apps.SelectOption{
					{Label: "1", Value: "1"},
					{Label: "2", Value: "2"},
					{Label: "3", Value: "3"},
					{Label: "4", Value: "4"},
					{Label: "5", Value: "5"},
				},
			},
			{
				Label: "Have you liked it?",
				Name:  fieldResponse,
				Type:  apps.FieldTypeButton,
				SelectStaticOptions: []apps.SelectOption{
					{Label: "Like", Value: "like"},
					{Label: "Dislike", Value: "dislike"},
				},
				TextSubtype: "submit",
			},
		},
	}
}

func (h *helloapp) fSurvey(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, c *apps.Call) (int, error) {
	var out *apps.CallResponse

	// userID := c.GetValue(fieldUserID, c.Context.ActingUserID)
	message := c.GetValue(fieldMessage, "default hello message")

	switch c.Type {
	case apps.CallTypeForm:
		out = h.newSurveyFormResponse(message)

	case apps.CallTypeSubmit:
		// TODO post something somewhere; for embedded form - what do we do?
		out = &apps.CallResponse{}
	}

	httputils.WriteJSON(w, out)
	return http.StatusOK, nil
}

func (h *helloapp) newSurveyFormResponse(message string) *apps.CallResponse {
	return &apps.CallResponse{
		Type: apps.CallResponseTypeForm,
		Form: h.newSurveyForm(message),
	}
}
