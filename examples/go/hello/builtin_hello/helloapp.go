package builtin_hello

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/pkg/errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/examples/go/hello"
	"github.com/mattermost/mattermost-plugin-apps/upstream"
	"github.com/mattermost/mattermost-plugin-apps/utils"
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

func New(mm *pluginapi.Client, log utils.Logger) *helloapp {
	return &helloapp{
		HelloApp: hello.NewHelloApp(mm, log),
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

func (h *helloapp) Roundtrip(_ *apps.App, c *apps.CallRequest, _ bool) (io.ReadCloser, error) {
	cr := &apps.CallResponse{}
	switch c.Path {
	case apps.DefaultBindings.Path:
		cr = &apps.CallResponse{
			Type: apps.CallResponseTypeOK,
			Data: hello.Bindings(),
		}

	case hello.PathInstall:
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
	return io.NopCloser(bytes.NewReader(bb)), nil
}

func (h *helloapp) GetStatic(_ *apps.Manifest, path string) (io.ReadCloser, int, error) {
	return nil, http.StatusNotFound, utils.ErrNotFound
}

func (h *helloapp) mapFunctions(c *apps.CallRequest) (*apps.CallResponse, error) {
	if strings.HasPrefix(c.Path, hello.PathSendSurveyModal) {
		return h.SendSurveyModal(c), nil
	}

	if strings.HasPrefix(c.Path, hello.PathSendSurveyCommandToModal) {
		callType := strings.TrimPrefix(c.Path, hello.PathSendSurveyCommandToModal+"/")
		return h.SendSurveyCommandToModal(c, apps.CallType(callType)), nil
	}

	if strings.HasPrefix(c.Path, hello.PathSendSurvey) {
		callType := strings.TrimPrefix(c.Path, hello.PathSendSurvey+"/")
		return h.SendSurvey(c, apps.CallType(callType)), nil
	}

	if strings.HasPrefix(c.Path, hello.PathSurvey) {
		callType := strings.TrimPrefix(c.Path, hello.PathSurvey+"/")
		return h.Survey(c, apps.CallType(callType)), nil
	}

	// notifications
	if strings.HasPrefix(c.Path, hello.PathUserJoinedChannel) {
		h.HelloApp.UserJoinedChannel(c)
		return &apps.CallResponse{Type: apps.CallResponseTypeOK}, nil
	}

	return nil, errors.Errorf("%s is not found", c.Path)
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
			Markdown: "<>/<> TODO",
		}
	default:
		return apps.NewErrorCallResponse(errors.Errorf("Unexpected call type: \"%s\"", callType))
	}
}
