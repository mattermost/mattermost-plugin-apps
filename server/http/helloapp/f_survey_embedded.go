package helloapp

import (
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
	"github.com/mattermost/mattermost-server/v5/model"
)

func (h *helloapp) fSurveyEmbedded(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, c *apps.Call) (int, error) {
	message := c.GetValue(fieldMessage, "default hello message")
	post := h.newSurveyFormPost(message)

	out := &apps.CallResponse{
		Type: apps.CallResponseTypeOK,
	}

	_, err := h.dmPost(c.Context.ActingUserID, post)
	if err != nil {
		out.Type = apps.CallResponseTypeError
		out.Error = fmt.Sprintf("Could not send the survey. Error: %s", err.Error())
	}
	httputils.WriteJSON(w, out)
	return http.StatusOK, nil
}

func (h *helloapp) fSurveyEmbeddedSubmit(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, c *apps.Call) (int, error) {
	response := apps.CallResponse{
		Type: apps.CallResponseTypeUpdateEmbedded,
		Data: make(map[string]interface{}),
	}
	post := &model.Post{
		Message: "Submitted",
		Props:   model.StringInterface{},
	}
	response.Data[apps.EmbeddedResponseDataPost] = post

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

func (h *helloapp) newSurveyFormPost(message string) *model.Post {
	form := h.newSurveyForm(message)
	call := h.makeCall(PathEmbedSurveySubmit)
	call.Context = &apps.Context{
		AppID: AppID,
	}
	return &model.Post{
		Message: "Fill this survey",
		Props: model.StringInterface{
			"form": form,
			"call": call,
		},
	}
}
