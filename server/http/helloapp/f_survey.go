package helloapp

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
	"github.com/mattermost/mattermost-server/v5/model"
)

func (h *helloapp) Survey(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, c *api.Call) (int, error) {
	var out *api.CallResponse

	// userID := c.GetValue(fieldUserID, c.Context.ActingUserID)
	message := c.GetValue(fieldMessage, "default hello message")

	switch c.Type {
	case api.CallTypeForm:
		out = &api.CallResponse{
			Type: api.CallResponseTypeForm,
			Form: h.newSurveyForm(message),
		}

	case api.CallTypeSubmit:
		// TODO post something somewhere; for embedded form - what do we do?
		out = &api.CallResponse{}
	}

	httputils.WriteJSON(w, out)
	return http.StatusOK, nil
}

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

func (h *helloapp) newSurveyDebugDialog(message string) *model.OpenDialogRequest {
	return &model.OpenDialogRequest{
		TriggerId: appID,
		URL:       h.appURL(pathSendSurveyDebugDialogSubmit),
		Dialog: model.Dialog{
			IconURL: "http://www.mattermost.org/wp-content/uploads/2016/04/icon.png",
			Elements: []model.DialogElement{
				{
					DisplayName: "Response",
					Name:        fieldResponse,
					Type:        "select",
					Options: []*model.PostActionOptions{
						{Text: "Like", Value: "like"},
						{Text: "Dislike", Value: "dislike"},
					},
				},
			},
		},
	}
}
