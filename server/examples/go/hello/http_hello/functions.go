package http_hello

import (
	"errors"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/examples/go/hello"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

func (h *helloapp) GetBindings(w http.ResponseWriter, req *http.Request, claims *api.JWTClaims, c *api.Call) (int, error) {
	httputils.WriteJSON(w, api.CallResponse{
		Type: api.CallResponseTypeOK,
		Data: hello.Bindings(),
	})
	return http.StatusOK, nil
}

func (h *helloapp) Install(w http.ResponseWriter, req *http.Request, claims *api.JWTClaims, c *api.Call) (int, error) {
	if c.Type != api.CallTypeSubmit {
		return http.StatusBadRequest, errors.New("not supported")
	}
	out, err := h.HelloApp.Install(AppID, AppDisplayName, c)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	httputils.WriteJSON(w, &api.CallResponse{
		Type:     api.CallResponseTypeOK,
		Markdown: out,
	})
	return http.StatusOK, nil
}

func (h *helloapp) SendSurvey(w http.ResponseWriter, req *http.Request, claims *api.JWTClaims, c *api.Call) (int, error) {
	var out *api.CallResponse
	switch c.Type {
	case api.CallTypeForm:
		out = hello.NewSendSurveyFormResponse(c)

	case api.CallTypeSubmit:
		txt, err := h.HelloApp.SendSurvey(c)
		if err != nil {
			return http.StatusInternalServerError, err
		}
		out = &api.CallResponse{
			Type:     api.CallResponseTypeOK,
			Markdown: txt,
		}
	case api.CallTypeLookup:
		out = &api.CallResponse{
			Data: map[string]interface{}{
				"items": []*api.SelectOption{
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

func (h *helloapp) SendSurveyModal(w http.ResponseWriter, req *http.Request, claims *api.JWTClaims, c *api.Call) (int, error) {
	out := hello.NewSendSurveyFormResponse(c)
	httputils.WriteJSON(w, out)
	return http.StatusOK, nil
}

func (h *helloapp) SendSurveyCommandToModal(w http.ResponseWriter, req *http.Request, claims *api.JWTClaims, c *api.Call) (int, error) {
	var out *api.CallResponse

	switch c.Type {
	case api.CallTypeSubmit:
		out = hello.NewSendSurveyFormResponse(c)
	default:
		out = hello.NewSendSurveyPartialFormResponse(c)
	}

	httputils.WriteJSON(w, out)
	return http.StatusOK, nil
}

func (h *helloapp) Survey(w http.ResponseWriter, req *http.Request, claims *api.JWTClaims, c *api.Call) (int, error) {
	var out *api.CallResponse

	switch c.Type {
	case api.CallTypeForm:
		out = hello.NewSurveyFormResponse(c)

	case api.CallTypeSubmit:
		err := h.ProcessSurvey(c)
		if err != nil {
			return http.StatusInternalServerError, err
		}
		out = &api.CallResponse{
			Type:     api.CallResponseTypeOK,
			Markdown: "<><> TODO",
		}
	}

	httputils.WriteJSON(w, out)
	return http.StatusOK, nil
}

func (h *helloapp) UserJoinedChannel(_ http.ResponseWriter, _ *http.Request, _ *api.JWTClaims, call *api.Call) (int, error) {
	h.HelloApp.UserJoinedChannel(call)
	return http.StatusOK, nil
}
