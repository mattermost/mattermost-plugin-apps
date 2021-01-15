package hello

import (
	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-server/v5/model"
)

func NewSurveyForm(message string) *api.Form {
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

func NewSurveyFormResponse(c *api.Call) *api.CallResponse {
	message := c.GetValue(fieldMessage, "default hello message")
	return &api.CallResponse{
		Type: api.CallResponseTypeForm,
		Form: NewSurveyForm(message),
	}
}

func (h *HelloApp) ProcessSurvey(c *api.Call) *api.CallResponse {
	if api.LocationInPost.In(c.Context.Location) {
		return &api.CallResponse{
			Type: api.CallResponseTypeUpdateEmbedded,
			Data: &model.Post{
				Message: "Survey submitted. Thanks!",
			},
		}
	}
	return &api.CallResponse{
		Type:     api.CallResponseTypeOK,
		Markdown: "Survey submitted. Thanks!",
	}
}
