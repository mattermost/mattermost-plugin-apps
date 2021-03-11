package builtin_hello

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"strings"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/examples/go/hello"
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

func Manifest() *apps.Manifest {
	return &apps.Manifest{
		AppID:       AppID,
		Type:        apps.AppTypeBuiltin,
		DisplayName: AppDisplayName,
		Description: AppDescription,
		RequestedPermissions: apps.Permissions{
			apps.PermissionUserJoinedChannelNotification,
			apps.PermissionActAsUser,
			apps.PermissionActAsBot,
		},
		RequestedLocations: apps.Locations{
			apps.LocationChannelHeader,
			apps.LocationPostMenu,
			apps.LocationCommand,
			apps.LocationInPost,
		},
		HomepageURL: ("https://github.com/mattermost"),
	}
}

func (h *helloapp) Roundtrip(c *apps.CallRequest) (io.ReadCloser, error) {
	cr := &apps.CallResponse{}
	switch c.Path {
	case api.BindingsPath:
		cr = &apps.CallResponse{
			Type: apps.CallResponseTypeOK,
			Data: hello.Bindings(),
		}

	case apps.DefaultInstallCallPath:
		cr = h.Install(c)
	default:
		var err error
		cr, err = h.mapFunctions(c)
		if err != nil {
			return nil, err
		}
	}

	bb, err := json.Marshal(cr)
	if err != nil {
		return nil, err
	}
	return ioutil.NopCloser(bytes.NewReader(bb)), nil
}

func (h *helloapp) mapFunctions(c *apps.CallRequest) (*apps.CallResponse, error) {
	if strings.HasPrefix(c.Path, hello.PathSendSurvey) {
		callType := strings.TrimPrefix(c.Path, hello.PathSendSurvey+"/")
		return h.SendSurvey(c, apps.CallType(callType)), nil
	}

	if strings.HasPrefix(c.Path, hello.PathSendSurveyModal) {
		return h.SendSurveyModal(c), nil
	}

	if strings.HasPrefix(c.Path, hello.PathSendSurveyCommandToModal) {
		callType := strings.TrimPrefix(c.Path, hello.PathSendSurveyCommandToModal+"/")
		return h.SendSurveyCommandToModal(c, apps.CallType(callType)), nil
	}

	if strings.HasPrefix(c.Path, hello.PathSurvey) {
		callType := strings.TrimPrefix(c.Path, hello.PathSurvey+"/")
		return h.Survey(c, apps.CallType(callType)), nil
	}

	return nil, errors.Errorf("%s is not found", c.Path)
}

func (h *helloapp) OneWay(call *apps.CallRequest) error {
	switch call.Context.Subject {
	case apps.SubjectUserJoinedChannel:
		h.HelloApp.UserJoinedChannel(call)
	default:
		return errors.Errorf("%s is not supported", call.Context.Subject)
	}
	return nil
}

func (h *helloapp) Install(c *apps.CallRequest) *apps.CallResponse {
	out, err := h.HelloApp.Install(AppID, AppDisplayName, c)
	if err != nil {
		return apps.NewErrorCallResponse(err)
	}
	return &apps.CallResponse{
		Type:     apps.CallResponseTypeOK,
		Markdown: out,
	}
}

func (h *helloapp) SendSurvey(c *apps.CallRequest, callType apps.CallType) *apps.CallResponse {
	switch callType {
	case apps.CallTypeForm:
		return hello.NewSendSurveyFormResponse(c)

	case apps.CallTypeSubmit:
		txt, err := h.HelloApp.SendSurvey(c)
		if err != nil {
			return apps.NewErrorCallResponse(err)
		}
		return &apps.CallResponse{
			Type:     apps.CallResponseTypeOK,
			Markdown: txt,
		}
	case apps.CallTypeLookup:
		return &apps.CallResponse{
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
		return apps.NewErrorCallResponse(errors.Errorf("Unexpected call type: \"%s\"", callType))
	}
}

func (h *helloapp) SendSurveyModal(c *apps.CallRequest) *apps.CallResponse {
	return hello.NewSendSurveyFormResponse(c)
}

func (h *helloapp) SendSurveyCommandToModal(c *apps.CallRequest, callType apps.CallType) *apps.CallResponse {
	return hello.NewSendSurveyPartialFormResponse(c, callType)
}

func (h *helloapp) Survey(c *apps.CallRequest, callType apps.CallType) *apps.CallResponse {
	switch callType {
	case apps.CallTypeForm:
		return hello.NewSurveyFormResponse(c)

	case apps.CallTypeSubmit:
		err := h.ProcessSurvey(c)
		if err != nil {
			return apps.NewErrorCallResponse(err)
		}
		return &apps.CallResponse{
			Type:     apps.CallResponseTypeOK,
			Markdown: "<><> TODO",
		}
	default:
		return apps.NewErrorCallResponse(errors.Errorf("Unexpected call type: \"%s\"", callType))
	}
}
