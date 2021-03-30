package http_hello

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/examples/go/hello"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
	"github.com/pkg/errors"
)

func (h *helloapp) GetBindings(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, c *apps.CallRequest) (int, error) {
	httputils.WriteJSON(w, apps.CallResponse{
		Type: apps.CallResponseTypeOK,
		Data: hello.Bindings(),
	})
	return http.StatusOK, nil
}

func (h *helloapp) Install(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, c *apps.CallRequest) (int, error) {
	out, err := h.HelloApp.Install(AppID, AppDisplayName, c)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	httputils.WriteJSON(w, &apps.CallResponse{
		Type:     apps.CallResponseTypeOK,
		Markdown: out,
	})
	return http.StatusOK, nil
}

func (h *helloapp) SendSurvey(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, c *apps.CallRequest) (int, error) {
	var out *apps.CallResponse
	vars := mux.Vars(req)
	callType := apps.CallType(vars["type"])

	switch callType {
	case apps.CallTypeForm:
		out = hello.NewSendSurveyFormResponse(c)

	case apps.CallTypeSubmit:
		txt, err := h.HelloApp.SendSurvey(c)
		if err != nil {
			return http.StatusInternalServerError, err
		}
		out = &apps.CallResponse{
			Type:     apps.CallResponseTypeOK,
			Markdown: txt,
		}
	case apps.CallTypeLookup:
		out = &apps.CallResponse{
			Data: map[string]interface{}{
				"items": []*apps.SelectOption{
					{
						Label: "Option 1",
						Value: "option1",
					},
				},
			},
		}
	default:
		out = apps.NewErrorCallResponse(errors.Errorf("Unexpected call type: \"%s\"", callType))
	}

	httputils.WriteJSON(w, out)
	return http.StatusOK, nil
}

func (h *helloapp) SendSurveyModal(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, c *apps.CallRequest) (int, error) {
	out := hello.NewSendSurveyFormResponse(c)
	httputils.WriteJSON(w, out)
	return http.StatusOK, nil
}

func (h *helloapp) SubmitSurvey(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, c *apps.CallRequest) (int, error) {
	out := hello.SubmitSurvey(c)
	httputils.WriteJSON(w, out)
	return http.StatusOK, nil
}

func (h *helloapp) SendSurveyCommandToModal(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, c *apps.CallRequest) (int, error) {
	var out *apps.CallResponse
	vars := mux.Vars(req)
	callType := apps.CallType(vars["type"])

	switch callType {
	case apps.CallTypeSubmit:
		out = hello.NewSendSurveyFormResponse(c)
	default:
		out = hello.NewSendSurveyPartialFormResponse(c, callType)
	}

	httputils.WriteJSON(w, out)
	return http.StatusOK, nil
}

func (h *helloapp) Survey(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, c *apps.CallRequest) (int, error) {
	var out *apps.CallResponse
	vars := mux.Vars(req)
	callType := apps.CallType(vars["type"])

	switch callType {
	case apps.CallTypeForm:
		out = hello.NewSurveyFormResponse(c)

	case apps.CallTypeSubmit:
		err := h.ProcessSurvey(c)
		if err != nil {
			return http.StatusInternalServerError, err
		}
		out = &apps.CallResponse{
			Type:     apps.CallResponseTypeOK,
			Markdown: "<><> TODO",
		}
	default:
		out = apps.NewErrorCallResponse(errors.Errorf("Unexpected call type: \"%s\"", callType))
	}

	httputils.WriteJSON(w, out)
	return http.StatusOK, nil
}

func (h *helloapp) UserJoinedChannel(_ http.ResponseWriter, _ *http.Request, _ *apps.JWTClaims, call *apps.CallRequest) (int, error) {
	h.HelloApp.UserJoinedChannel(call)
	return http.StatusOK, nil
}
