package helloapp

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

func (h *helloapp) newSurveyForm(message string) *apps.Form {
	return &apps.Form{
		Title:         "Emotional response survey",
		Header:        message,
		Footer:        "Let the world know!",
		SubmitButtons: fieldResponse,
		Call:          h.makeCall(PathSendSurvey),
		Fields: []*apps.Field{
			{
				Name: fieldResponse,
				Type: apps.FieldTypeStaticSelect,
				SelectStaticOptions: []apps.SelectOption{
					{Label: "Like", Value: "like"},
					{Label: "Dislike", Value: "dislike"},
				},
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
