package builtin_hello

import (
	"github.com/mattermost/mattermost-plugin-apps/server/api"
)

func Survey(c *api.Call) *api.CallResponse {
	message := c.GetValue(fieldMessage, "default hello message")
	switch c.Type {
	case api.CallTypeForm:
		return newSurveyFormResponse(message)

	case api.CallTypeSubmit:
		// TODO post something somewhere; for embedded form - what do we do?
		return &api.CallResponse{}
	}

	return nil
}

func newSurveyForm(message string) *api.Form {
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

func newSurveyFormResponse(message string) *api.CallResponse {
	return &api.CallResponse{
		Type: api.CallResponseTypeForm,
		Form: newSurveyForm(message),
	}
}
