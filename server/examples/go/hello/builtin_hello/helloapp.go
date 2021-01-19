package builtin_hello

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"

	"github.com/mattermost/mattermost-plugin-apps/modelapps"
	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/examples/go/hello"
	"github.com/pkg/errors"
)

const (
	AppID          = "builtin"
	AppDisplayName = "builtin hello display name"
	AppDescription = "builtin hello description"
)

type helloapp struct {
	*hello.HelloApp
}

var _ api.Upstream = (*helloapp)(nil)

func New(appService *api.Service) *helloapp {
	return &helloapp{
		HelloApp: &hello.HelloApp{
			API: appService,
		},
	}
}

func Manifest() *modelapps.Manifest {
	return &modelapps.Manifest{
		AppID:       AppID,
		Type:        modelapps.AppTypeBuiltin,
		DisplayName: AppDisplayName,
		Description: AppDescription,
		RequestedPermissions: modelapps.Permissions{
			modelapps.PermissionUserJoinedChannelNotification,
			modelapps.PermissionActAsUser,
			modelapps.PermissionActAsBot,
		},
		RequestedLocations: modelapps.Locations{
			modelapps.LocationChannelHeader,
			modelapps.LocationPostMenu,
			modelapps.LocationCommand,
			modelapps.LocationInPost,
		},
		HomepageURL: ("https://github.com/mattermost"),
	}
}

func (h *helloapp) Roundtrip(c *modelapps.Call) (io.ReadCloser, error) {
	cr := &modelapps.CallResponse{}
	switch c.URL {
	case api.BindingsPath:
		cr = &modelapps.CallResponse{
			Type: modelapps.CallResponseTypeOK,
			Data: hello.Bindings(),
		}

	case modelapps.DefaultInstallCallPath:
		cr = h.Install(c)
	case hello.PathSendSurvey:
		cr = h.SendSurvey(c)
	case hello.PathSendSurveyModal:
		cr = h.SendSurveyModal(c)
	case hello.PathSendSurveyCommandToModal:
		cr = h.SendSurveyCommandToModal(c)
	case hello.PathSurvey:
		cr = h.Survey(c)
	default:
		return nil, errors.Errorf("%s is not found", c.URL)
	}

	bb, err := json.Marshal(cr)
	if err != nil {
		return nil, err
	}
	return ioutil.NopCloser(bytes.NewReader(bb)), nil
}

func (h *helloapp) OneWay(call *modelapps.Call) error {
	switch call.Context.Subject {
	case modelapps.SubjectUserJoinedChannel:
		h.HelloApp.UserJoinedChannel(call)
	default:
		return errors.Errorf("%s is not supported", call.Context.Subject)
	}
	return nil
}

func (h *helloapp) Install(c *modelapps.Call) *modelapps.CallResponse {
	if c.Type != modelapps.CallTypeSubmit {
		return modelapps.NewErrorCallResponse(errors.New("not supported"))
	}
	out, err := h.HelloApp.Install(AppID, AppDisplayName, c)
	if err != nil {
		return modelapps.NewErrorCallResponse(err)
	}
	return &modelapps.CallResponse{
		Type:     modelapps.CallResponseTypeOK,
		Markdown: out,
	}
}

func (h *helloapp) SendSurvey(c *modelapps.Call) *modelapps.CallResponse {
	switch c.Type {
	case modelapps.CallTypeForm:
		return hello.NewSendSurveyFormResponse(c)

	case modelapps.CallTypeSubmit:
		txt, err := h.HelloApp.SendSurvey(c)
		if err != nil {
			return modelapps.NewErrorCallResponse(err)
		}
		return &modelapps.CallResponse{
			Type:     modelapps.CallResponseTypeOK,
			Markdown: txt,
		}
	case modelapps.CallTypeLookup:
		return &modelapps.CallResponse{
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

	return nil
}

func (h *helloapp) SendSurveyModal(c *modelapps.Call) *modelapps.CallResponse {
	return hello.NewSendSurveyFormResponse(c)
}

func (h *helloapp) SendSurveyCommandToModal(c *modelapps.Call) *modelapps.CallResponse {
	return hello.NewSendSurveyPartialFormResponse(c)
}

func (h *helloapp) Survey(c *modelapps.Call) *modelapps.CallResponse {
	switch c.Type {
	case modelapps.CallTypeForm:
		return hello.NewSurveyFormResponse(c)

	case modelapps.CallTypeSubmit:
		err := h.ProcessSurvey(c)
		if err != nil {
			return modelapps.NewErrorCallResponse(err)
		}
		return &modelapps.CallResponse{
			Type:     modelapps.CallResponseTypeOK,
			Markdown: "<><> TODO",
		}
	}
	return nil
}
