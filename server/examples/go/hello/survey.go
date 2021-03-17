package hello

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func NewSurveyForm(message string) *apps.Form {
	return &apps.Form{
		Title:         "Emotional response survey",
		Header:        message,
		Footer:        "Let the world know!",
		SubmitButtons: fieldResponse,
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

func NewSurveyFormResponse(c *apps.CallRequest) *apps.CallResponse {
	message := c.GetValue(fieldMessage, "default hello message")
	return &apps.CallResponse{
		Type: apps.CallResponseTypeForm,
		Form: NewSurveyForm(message),
	}
}

func (h *HelloApp) ProcessSurvey(c *apps.CallRequest) error {
	// TODO post something; for embedded form - what do we do?
	return nil
}
