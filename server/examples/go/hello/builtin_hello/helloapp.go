package builtin_hello

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"

	"github.com/pkg/errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
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

func New(mm *pluginapi.Client) *helloapp {
	return &helloapp{
		HelloApp: hello.NewHelloApp(mm),
	}
}

func Manifest() *apps.Manifest {
	return &apps.Manifest{
		AppID:       AppID,
		Type:        apps.AppTypeBuiltin,
		Version:     "0.1.0",
		DisplayName: AppDisplayName,
		Description: AppDescription,
		HomepageURL: ("https://github.com/mattermost"),
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
	}
}

func App() *apps.App {
	m := *Manifest()
	m.Version = "pre-release"

	return &apps.App{
		Manifest:           m,
		GrantedPermissions: m.RequestedPermissions,
		GrantedLocations:   m.RequestedLocations,
	}
}

func (h *helloapp) Roundtrip(c *apps.Call) (io.ReadCloser, error) {
	cr := &apps.CallResponse{}
	switch c.Path {
	case api.BindingsPath:
		cr = &apps.CallResponse{
			Type: apps.CallResponseTypeOK,
			Data: hello.Bindings(),
		}

	case apps.DefaultInstallCallPath:
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
		return nil, errors.Errorf("%s is not found", c.Path)
	}

	bb, err := json.Marshal(cr)
	if err != nil {
		return nil, err
	}
	return ioutil.NopCloser(bytes.NewReader(bb)), nil
}

func (h *helloapp) OneWay(call *apps.Call) error {
	switch call.Context.Subject {
	case apps.SubjectUserJoinedChannel:
		h.HelloApp.UserJoinedChannel(call)
	default:
		return errors.Errorf("%s is not supported", call.Context.Subject)
	}
	return nil
}

func (h *helloapp) Install(c *apps.Call) *apps.CallResponse {
	if c.Type != apps.CallTypeSubmit {
		return apps.NewErrorCallResponse(errors.New("not supported"))
	}
	out, err := h.HelloApp.Install(AppID, AppDisplayName, c)
	if err != nil {
		return apps.NewErrorCallResponse(err)
	}
	return &apps.CallResponse{
		Type:     apps.CallResponseTypeOK,
		Markdown: out,
	}
}

func (h *helloapp) SendSurvey(c *apps.Call) *apps.CallResponse {
	switch c.Type {
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
	}

	return nil
}

func (h *helloapp) SendSurveyModal(c *apps.Call) *apps.CallResponse {
	return hello.NewSendSurveyFormResponse(c)
}

func (h *helloapp) SendSurveyCommandToModal(c *apps.Call) *apps.CallResponse {
	return hello.NewSendSurveyPartialFormResponse(c)
}

func (h *helloapp) Survey(c *apps.Call) *apps.CallResponse {
	switch c.Type {
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
	}
	return nil
}
