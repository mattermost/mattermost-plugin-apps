package hello

import (
	"github.com/mattermost/mattermost-plugin-apps/server/api"
)

func NewSurveyForm(message string) *api.Form {
	return &api.Form{
		Title:         "Emotional response survey",
		Header:        message,
		Footer:        "Let the world know!",
		SubmitButtons: fieldResponse,
		Call:          api.MakeCall(PathSendSurvey),
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

func NewSurveyFormResponse(c *api.Call) *api.CallResponse {
	message := c.GetValue(fieldMessage, "default hello message")
	return &api.CallResponse{
		Type: api.CallResponseTypeForm,
		Form: NewSurveyForm(message),
	}
}

func (h *HelloApp) ProcessSurvey(c *api.Call) error {
	// TODO post something; for embedded form - what do we do?
	return nil
}
