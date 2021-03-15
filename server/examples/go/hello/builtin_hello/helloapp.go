package builtin_hello

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/examples/go/hello"
	"github.com/mattermost/mattermost-plugin-apps/server/upstream"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
)

const (
	AppID          = "builtin"
	AppDisplayName = "builtin hello display name"
	AppDescription = "builtin hello description"
)

type helloapp struct {
	*hello.HelloApp
}

var _ upstream.Upstream = (*helloapp)(nil)

func New(mm *pluginapi.Client) *helloapp {
	return &helloapp{
		HelloApp: hello.NewHelloApp(mm),
	}
}

func Manifest() *apps.Manifest {
	return &apps.Manifest{
		AppID:       AppID,
		AppType:     apps.AppTypeBuiltin,
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

func (h *helloapp) Roundtrip(c *apps.CallRequest, _ bool) (io.ReadCloser, error) {
	cr := &apps.CallResponse{}
	switch c.Path {
	case apps.DefaultBindingsCallPath:
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
	case hello.PathUserJoinedChannel:
		h.HelloApp.UserJoinedChannel(c)
	default:
		return nil, errors.Errorf("%s is not found", c.Path)
	}

	bb, err := json.Marshal(cr)
	if err != nil {
		return nil, err
	}
	return ioutil.NopCloser(bytes.NewReader(bb)), nil
}

func (h *helloapp) GetStatic(path string) (io.ReadCloser, int, error) {
	return nil, http.StatusNotFound, utils.ErrNotFound
}

func (h *helloapp) Install(c *apps.CallRequest) *apps.CallResponse {
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

func (h *helloapp) SendSurvey(c *apps.CallRequest) *apps.CallResponse {
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
	default:
		return apps.NewErrorCallResponse(errors.Errorf("Unexpected call type: \"%s\"", c.Type))
	}
}

func (h *helloapp) SendSurveyModal(c *apps.CallRequest) *apps.CallResponse {
	return hello.NewSendSurveyFormResponse(c)
}

func (h *helloapp) SendSurveyCommandToModal(c *apps.CallRequest) *apps.CallResponse {
	return hello.NewSendSurveyPartialFormResponse(c)
}

func (h *helloapp) Survey(c *apps.CallRequest) *apps.CallResponse {
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
	default:
		return apps.NewErrorCallResponse(errors.Errorf("Unexpected call type: \"%s\"", c.Type))
	}
}
