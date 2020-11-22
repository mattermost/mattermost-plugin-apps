package http_hello

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

func (h *helloapp) newSurveyForm(message string) *api.Form {
	return &api.Form{
		Title:         "Emotional response survey",
		Header:        message,
		Footer:        "Let the world know!",
		SubmitButtons: fieldResponse,
		Fields: []*api.Field{
			{
				Name: fieldResponse,
				Type: api.FieldTypeStaticSelect,
				SelectStaticOptions: []api.SelectOption{
					{Label: "Like", Value: "like"},
					{Label: "Dislike", Value: "dislike"},
				},
			},
		},
	}
}

func (h *helloapp) fSurvey(w http.ResponseWriter, req *http.Request, claims *api.JWTClaims, c *api.Call) (int, error) {
	var out *api.CallResponse

	// userID := c.GetValue(fieldUserID, c.Context.ActingUserID)
	message := c.GetValue(fieldMessage, "default hello message")

	switch c.Type {
	case api.CallTypeForm:
		out = h.newSurveyFormResponse(message)

	case api.CallTypeSubmit:
		// TODO post something somewhere; for embedded form - what do we do?
		out = &api.CallResponse{}
	}

	httputils.WriteJSON(w, out)
	return http.StatusOK, nil
}

func (h *helloapp) newSurveyFormResponse(message string) *api.CallResponse {
	return &api.CallResponse{
		Type: api.CallResponseTypeForm,
		Form: h.newSurveyForm(message),
	}
}
