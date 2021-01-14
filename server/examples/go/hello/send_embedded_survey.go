package hello

import (
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/examples"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
	"github.com/mattermost/mattermost-server/v5/model"
)

func (h *HelloApp) fSurveyEmbedded(w http.ResponseWriter, req *http.Request, claims *api.JWTClaims, c *api.Call) (int, error) {
	bot := examples.AsBot(c.Context)
	message := c.GetValue(fieldMessage, "default hello message")
	post := h.newSurveyFormPost(message)

	out := &api.CallResponse{
		Type: api.CallResponseTypeOK,
	}

	_, err := bot.DMPost(c.Context.ActingUserID, post)
	if err != nil {
		out.Type = api.CallResponseTypeError
		out.ErrorText = fmt.Sprintf("Could not send the survey. Error: %s", err.Error())
	}
	httputils.WriteJSON(w, out)
	return http.StatusOK, nil
}

func (h *HelloApp) fSurveyEmbeddedSubmit(w http.ResponseWriter, req *http.Request, claims *api.JWTClaims, c *api.Call) (int, error) {
	response := api.CallResponse{
		Type: api.CallResponseTypeUpdateEmbedded,
	}
	post := &model.Post{
		Message: "Submitted",
		Props:   model.StringInterface{},
	}
	response.Data = map[string]interface{}{
		api.EmbeddedResponseDataPost: post,
	}

	// Here as error handling example
	// response := apps.CallResponse{
	// 	Type:  apps.ResponseTypeError,
	// 	Data:  make(map[string]interface{}),
	// 	Error: "Some error",
	// }

	// errors := map[string]string{}
	// for key := range data.Values.Data {
	// 	errors[key] = "Some other error"
	// }
	// response.Data[api.EmbeddedResponseDataErrors] = errors

	httputils.WriteJSON(w, response)
	return http.StatusOK, nil
}

func (h *HelloApp) newSurveyFormPost(message string) *model.Post {
	form := NewSurveyForm(message)
	call := api.MakeCall(PathEmbedSurveySubmit)
	call.Context = &api.Context{
		AppID: "http-hello",
	}
	return &model.Post{
		Message: "Fill this survey",
		Props: model.StringInterface{
			"form": form,
			"call": call,
		},
	}
}
