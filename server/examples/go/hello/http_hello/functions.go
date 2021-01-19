package http_hello

import (
	"errors"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/modelapps"
	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/examples/go/hello"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

func (h *helloapp) GetBindings(w http.ResponseWriter, req *http.Request, claims *api.JWTClaims, c *modelapps.Call) (int, error) {
	httputils.WriteJSON(w, modelapps.CallResponse{
		Type: modelapps.CallResponseTypeOK,
		Data: hello.Bindings(),
	})
	return http.StatusOK, nil
}

func (h *helloapp) Install(w http.ResponseWriter, req *http.Request, claims *api.JWTClaims, c *modelapps.Call) (int, error) {
	if c.Type != modelapps.CallTypeSubmit {
		return http.StatusBadRequest, errors.New("not supported")
	}
	out, err := h.HelloApp.Install(AppID, AppDisplayName, c)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	httputils.WriteJSON(w, &modelapps.CallResponse{
		Type:     modelapps.CallResponseTypeOK,
		Markdown: out,
	})
	return http.StatusOK, nil
}

func (h *helloapp) SendSurvey(w http.ResponseWriter, req *http.Request, claims *api.JWTClaims, c *modelapps.Call) (int, error) {
	var out *modelapps.CallResponse
	switch c.Type {
	case modelapps.CallTypeForm:
		out = hello.NewSendSurveyFormResponse(c)

	case modelapps.CallTypeSubmit:
		txt, err := h.HelloApp.SendSurvey(c)
		if err != nil {
			return http.StatusInternalServerError, err
		}
		out = &modelapps.CallResponse{
			Type:     modelapps.CallResponseTypeOK,
			Markdown: txt,
		}
	case modelapps.CallTypeLookup:
		out = &modelapps.CallResponse{
			Data: map[string]interface{}{
				"items": []*modelapps.SelectOption{
					{
						Label: "Option 1",
						Value: "option1",
					},
				},
			},
		}
	}

	httputils.WriteJSON(w, out)
	return http.StatusOK, nil
}

func (h *helloapp) SendSurveyModal(w http.ResponseWriter, req *http.Request, claims *api.JWTClaims, c *modelapps.Call) (int, error) {
	out := hello.NewSendSurveyFormResponse(c)
	httputils.WriteJSON(w, out)
	return http.StatusOK, nil
}

func (h *helloapp) SendSurveyCommandToModal(w http.ResponseWriter, req *http.Request, claims *api.JWTClaims, c *modelapps.Call) (int, error) {
	var out *modelapps.CallResponse

	switch c.Type {
	case modelapps.CallTypeSubmit:
		out = hello.NewSendSurveyFormResponse(c)
	default:
		out = hello.NewSendSurveyPartialFormResponse(c)
	}

	httputils.WriteJSON(w, out)
	return http.StatusOK, nil
}

func (h *helloapp) Survey(w http.ResponseWriter, req *http.Request, claims *api.JWTClaims, c *modelapps.Call) (int, error) {
	var out *modelapps.CallResponse

	switch c.Type {
	case modelapps.CallTypeForm:
		out = hello.NewSurveyFormResponse(c)

	case modelapps.CallTypeSubmit:
		err := h.ProcessSurvey(c)
		if err != nil {
			return http.StatusInternalServerError, err
		}
		out = &modelapps.CallResponse{
			Type:     modelapps.CallResponseTypeOK,
			Markdown: "<><> TODO",
		}
	}

	httputils.WriteJSON(w, out)
	return http.StatusOK, nil
}

func (h *helloapp) UserJoinedChannel(_ http.ResponseWriter, _ *http.Request, _ *api.JWTClaims, call *modelapps.Call) (int, error) {
	h.HelloApp.UserJoinedChannel(call)
	return http.StatusOK, nil
}
