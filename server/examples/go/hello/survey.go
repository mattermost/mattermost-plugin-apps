package hello

import (
	"github.com/mattermost/mattermost-plugin-apps/modelapps"
)

func NewSurveyForm(message string) *modelapps.Form {
	return &modelapps.Form{
		Title:         "Emotional response survey",
		Header:        message,
		Footer:        "Let the world know!",
		SubmitButtons: fieldResponse,
		Fields: []*modelapps.Field{
			{
				Name: fieldResponse,
				Type: modelapps.FieldTypeStaticSelect,
				SelectStaticOptions: []modelapps.SelectOption{
					{Label: "Like", Value: "like"},
					{Label: "Dislike", Value: "dislike"},
				},
			},
		},
	}
}

func NewSurveyFormResponse(c *modelapps.Call) *modelapps.CallResponse {
	message := c.GetValue(fieldMessage, "default hello message")
	return &modelapps.CallResponse{
		Type: modelapps.CallResponseTypeForm,
		Form: NewSurveyForm(message),
	}
}

func (h *HelloApp) ProcessSurvey(c *modelapps.Call) error {
	// TODO post something; for embedded form - what do we do?
	return nil
}
