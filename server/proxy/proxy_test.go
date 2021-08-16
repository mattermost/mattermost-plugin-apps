package proxy

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/mocks/mock_store"
	"github.com/mattermost/mattermost-plugin-apps/server/mocks/mock_upstream"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/upstream"
)

func TestAppMetadataForClient(t *testing.T) {
	testApps := []*apps.App{
		{
			BotUserID:   "botid",
			BotUsername: "botusername",
			Manifest: apps.Manifest{
				AppID:       apps.AppID("app1"),
				AppType:     apps.AppTypeBuiltin,
				DisplayName: "App 1",
			},
		},
	}

	ctrl := gomock.NewController(t)
	p := newTestProxy(testApps, ctrl, nil, nil)
	c := &apps.CallRequest{
		Context: &apps.Context{
			UserAgentContext: apps.UserAgentContext{
				AppID: "app1",
			},
		},
		Call: apps.Call{
			Path: "/",
		},
	}

	resp := p.Call("session_id", "acting_user_id", c)
	require.Equal(t, resp.AppMetadata, &apps.AppMetadataForClient{
		BotUserID:   "botid",
		BotUsername: "botusername",
	})
}

func TestFormFilter(t *testing.T) {
	testApps := []*apps.App{
		{
			BotUserID:   "botid",
			BotUsername: "botusername",
			Manifest: apps.Manifest{
				AppID:       apps.AppID("app1"),
				AppType:     apps.AppTypeBuiltin,
				DisplayName: "App 1",
			},
		},
	}

	type TC = struct {
		name            string
		appResponse     *apps.CallResponse
		result          *apps.CallResponse
		apiExpectations func(api *plugintest.API)
	}
	testCases := []TC{
		{
			name: "no field filter on names",
			appResponse: &apps.CallResponse{
				Type: apps.CallResponseTypeForm,
				Form: &apps.Form{
					Title: "Test",
					Call: &apps.Call{
						Path: "/url",
					},
					Fields: []*apps.Field{
						{
							Name: "field1",
						},
						{
							Name: "field2",
						},
					},
				},
			},
			result: &apps.CallResponse{
				Type: apps.CallResponseTypeForm,
				Form: &apps.Form{
					Title: "Test",
					Call: &apps.Call{
						Path: "/url",
					},
					Fields: []*apps.Field{
						{
							Name: "field1",
						},
						{
							Name: "field2",
						},
					},
				},
			},
			apiExpectations: nil,
		},
		{
			name: "no field filter on labels",
			appResponse: &apps.CallResponse{
				Type: apps.CallResponseTypeForm,
				Form: &apps.Form{
					Title: "Test",
					Call: &apps.Call{
						Path: "/url",
					},
					Fields: []*apps.Field{
						{
							Label: "field1",
							Name:  "same",
						},
						{
							Label: "field2",
							Name:  "same",
						},
					},
				},
			},
			result: &apps.CallResponse{
				Type: apps.CallResponseTypeForm,
				Form: &apps.Form{
					Title: "Test",
					Call: &apps.Call{
						Path: "/url",
					},
					Fields: []*apps.Field{
						{
							Label: "field1",
							Name:  "same",
						},
						{
							Label: "field2",
							Name:  "same",
						},
					},
				},
			},
			apiExpectations: nil,
		},
		{
			name: "field filter with no name",
			appResponse: &apps.CallResponse{
				Type: apps.CallResponseTypeForm,
				Form: &apps.Form{
					Title: "Test",
					Call: &apps.Call{
						Path: "/url",
					},
					Fields: []*apps.Field{
						{
							Label: "field1",
						},
						{
							Label: "field2",
							Name:  "same",
						},
					},
				},
			},
			result: &apps.CallResponse{
				Type: apps.CallResponseTypeForm,
				Form: &apps.Form{
					Title: "Test",
					Call: &apps.Call{
						Path: "/url",
					},
					Fields: []*apps.Field{
						{
							Label: "field2",
							Name:  "same",
						},
					},
				},
			},
			apiExpectations: func(api *plugintest.API) {
				api.On("LogDebug", "App from malformed: Field with no name", "field", &apps.Field{
					Label: "field1",
				}).Times(1)
			},
		},
		{
			name: "field filter with same label inferred from name",
			appResponse: &apps.CallResponse{
				Type: apps.CallResponseTypeForm,
				Form: &apps.Form{
					Title: "Test",
					Call: &apps.Call{
						Path: "/url",
					},
					Fields: []*apps.Field{
						{
							Type: apps.FieldTypeBool,
							Name: "same",
						},
						{
							Type: apps.FieldTypeChannel,
							Name: "same",
						},
					},
				},
			},
			result: &apps.CallResponse{
				Type: apps.CallResponseTypeForm,
				Form: &apps.Form{
					Title: "Test",
					Call: &apps.Call{
						Path: "/url",
					},
					Fields: []*apps.Field{
						{
							Type: apps.FieldTypeBool,
							Name: "same",
						},
					},
				},
			},
			apiExpectations: func(api *plugintest.API) {
				api.On("LogDebug", "App from malformed: Field label repeated. Only getting first field with the label.", "label", "same").Times(1)
			},
		},
		{
			name: "field filter with same label",
			appResponse: &apps.CallResponse{
				Type: apps.CallResponseTypeForm,
				Form: &apps.Form{
					Title: "Test",
					Call: &apps.Call{
						Path: "/url",
					},
					Fields: []*apps.Field{
						{
							Type:  apps.FieldTypeBool,
							Label: "same",
							Name:  "field1",
						},
						{
							Type:  apps.FieldTypeChannel,
							Label: "same",
							Name:  "field2",
						},
					},
				},
			},
			result: &apps.CallResponse{
				Type: apps.CallResponseTypeForm,
				Form: &apps.Form{
					Title: "Test",
					Call: &apps.Call{
						Path: "/url",
					},
					Fields: []*apps.Field{
						{
							Type:  apps.FieldTypeBool,
							Label: "same",
							Name:  "field1",
						},
					},
				},
			},
			apiExpectations: func(api *plugintest.API) {
				api.On("LogDebug", "App from malformed: Field label repeated. Only getting first field with the label.", "label", "same").Times(1)
			},
		},
		{
			name: "field filter with same label",
			appResponse: &apps.CallResponse{
				Type: apps.CallResponseTypeForm,
				Form: &apps.Form{
					Title: "Test",
					Call: &apps.Call{
						Path: "/url",
					},
					Fields: []*apps.Field{
						{
							Type:  apps.FieldTypeBool,
							Label: "same",
							Name:  "field1",
						},
						{
							Type:  apps.FieldTypeChannel,
							Label: "same",
							Name:  "field2",
						},
					},
				},
			},
			result: &apps.CallResponse{
				Type: apps.CallResponseTypeForm,
				Form: &apps.Form{
					Title: "Test",
					Call: &apps.Call{
						Path: "/url",
					},
					Fields: []*apps.Field{
						{
							Type:  apps.FieldTypeBool,
							Label: "same",
							Name:  "field1",
						},
					},
				},
			},
			apiExpectations: func(api *plugintest.API) {
				api.On("LogDebug", "App from malformed: Field label repeated. Only getting first field with the label.", "label", "same").Times(1)
			},
		},
		{
			name: "field filter with multiword name",
			appResponse: &apps.CallResponse{
				Type: apps.CallResponseTypeForm,
				Form: &apps.Form{
					Title: "Test",
					Call: &apps.Call{
						Path: "/url",
					},
					Fields: []*apps.Field{
						{
							Type:  apps.FieldTypeBool,
							Label: "multiple word",
							Name:  "multiple word",
						},
					},
				},
			},
			result: &apps.CallResponse{
				Type: apps.CallResponseTypeForm,
				Form: &apps.Form{
					Title: "Test",
					Call: &apps.Call{
						Path: "/url",
					},
					Fields: []*apps.Field{},
				},
			},
			apiExpectations: func(api *plugintest.API) {
				api.On("LogDebug", "App form malformed: Name must be a single word", "name", "multiple word").Times(1)
			},
		},
		{
			name: "field filter with multiword label",
			appResponse: &apps.CallResponse{
				Type: apps.CallResponseTypeForm,
				Form: &apps.Form{
					Title: "Test",
					Call: &apps.Call{
						Path: "/url",
					},
					Fields: []*apps.Field{
						{
							Type:  apps.FieldTypeBool,
							Label: "multiple word",
							Name:  "singleword",
						},
					},
				},
			},
			result: &apps.CallResponse{
				Type: apps.CallResponseTypeForm,
				Form: &apps.Form{
					Title: "Test",
					Call: &apps.Call{
						Path: "/url",
					},
					Fields: []*apps.Field{},
				},
			},
			apiExpectations: func(api *plugintest.API) {
				api.On("LogDebug", "App form malformed: Label must be a single word", "label", "multiple word").Times(1)
			},
		},
		{
			name: "field filter more than one field",
			appResponse: &apps.CallResponse{
				Type: apps.CallResponseTypeForm,
				Form: &apps.Form{
					Title: "Test",
					Call: &apps.Call{
						Path: "/url",
					},
					Fields: []*apps.Field{
						{
							Type:  apps.FieldTypeBool,
							Label: "same",
							Name:  "field1",
						},
						{
							Type:  apps.FieldTypeChannel,
							Label: "same",
							Name:  "field2",
						},
						{
							Type:  apps.FieldTypeText,
							Label: "same",
							Name:  "field2",
						},
					},
				},
			},
			result: &apps.CallResponse{
				Type: apps.CallResponseTypeForm,
				Form: &apps.Form{
					Title: "Test",
					Call: &apps.Call{
						Path: "/url",
					},
					Fields: []*apps.Field{
						{
							Type:  apps.FieldTypeBool,
							Label: "same",
							Name:  "field1",
						},
					},
				},
			},
			apiExpectations: func(api *plugintest.API) {
				api.On("LogDebug", "App from malformed: Field label repeated. Only getting first field with the label.", "label", "same").Times(2)
			},
		},
		{
			name: "field filter static with no options",
			appResponse: &apps.CallResponse{
				Type: apps.CallResponseTypeForm,
				Form: &apps.Form{
					Title: "Test",
					Call: &apps.Call{
						Path: "/url",
					},
					Fields: []*apps.Field{
						{
							Type: apps.FieldTypeStaticSelect,
							Name: "field1",
						},
					},
				},
			},
			result: &apps.CallResponse{
				Type: apps.CallResponseTypeForm,
				Form: &apps.Form{
					Title: "Test",
					Call: &apps.Call{
						Path: "/url",
					},
					Fields: []*apps.Field{},
				},
			},
			apiExpectations: func(api *plugintest.API) {
				api.On("LogDebug", "App from malformed: Static field without opions.", "label", "field1").Times(1)
			},
		},
		{
			name: "field filter static options with no label",
			appResponse: &apps.CallResponse{
				Type: apps.CallResponseTypeForm,
				Form: &apps.Form{
					Title: "Test",
					Call: &apps.Call{
						Path: "/url",
					},
					Fields: []*apps.Field{
						{
							Type: apps.FieldTypeStaticSelect,
							Name: "field1",
							SelectStaticOptions: []apps.SelectOption{
								{
									Value: "opt1",
								},
								{},
							},
						},
					},
				},
			},
			result: &apps.CallResponse{
				Type: apps.CallResponseTypeForm,
				Form: &apps.Form{
					Title: "Test",
					Call: &apps.Call{
						Path: "/url",
					},
					Fields: []*apps.Field{
						{
							Type: apps.FieldTypeStaticSelect,
							Name: "field1",
							SelectStaticOptions: []apps.SelectOption{
								{
									Value: "opt1",
								},
							},
						},
					},
				},
			},
			apiExpectations: func(api *plugintest.API) {
				api.On("LogDebug", "App from malformed: Option with no label", "field", &apps.Field{
					Type: apps.FieldTypeStaticSelect,
					Name: "field1",
					SelectStaticOptions: []apps.SelectOption{
						{
							Value: "opt1",
						},
						{},
					},
				}, "option value", "").Times(1)
			},
		},
		{
			name: "field filter static options with same label inferred from value",
			appResponse: &apps.CallResponse{
				Type: apps.CallResponseTypeForm,
				Form: &apps.Form{
					Title: "Test",
					Call: &apps.Call{
						Path: "/url",
					},
					Fields: []*apps.Field{
						{
							Type: apps.FieldTypeStaticSelect,
							Name: "field1",
							SelectStaticOptions: []apps.SelectOption{
								{
									Value:    "same",
									IconData: "opt1",
								},
								{
									Value:    "same",
									IconData: "opt2",
								},
							},
						},
					},
				},
			},
			result: &apps.CallResponse{
				Type: apps.CallResponseTypeForm,
				Form: &apps.Form{
					Title: "Test",
					Call: &apps.Call{
						Path: "/url",
					},
					Fields: []*apps.Field{
						{
							Type: apps.FieldTypeStaticSelect,
							Name: "field1",
							SelectStaticOptions: []apps.SelectOption{
								{
									Value:    "same",
									IconData: "opt1",
								},
							},
						},
					},
				},
			},
			apiExpectations: func(api *plugintest.API) {
				api.On("LogDebug", "App from malformed: Repeated label on select option. Only getting first value with the label", "field", &apps.Field{
					Type: apps.FieldTypeStaticSelect,
					Name: "field1",
					SelectStaticOptions: []apps.SelectOption{
						{
							Value:    "same",
							IconData: "opt1",
						},
						{
							Value:    "same",
							IconData: "opt2",
						},
					},
				}, "option", apps.SelectOption{
					Value:    "same",
					IconData: "opt2",
				}).Times(1)
			},
		},
		{
			name: "field filter static options with same label",
			appResponse: &apps.CallResponse{
				Type: apps.CallResponseTypeForm,
				Form: &apps.Form{
					Title: "Test",
					Call: &apps.Call{
						Path: "/url",
					},
					Fields: []*apps.Field{
						{
							Type: apps.FieldTypeStaticSelect,
							Name: "field1",
							SelectStaticOptions: []apps.SelectOption{
								{
									Label: "same",
									Value: "opt1",
								},
								{
									Label: "same",
									Value: "opt2",
								},
							},
						},
					},
				},
			},
			result: &apps.CallResponse{
				Type: apps.CallResponseTypeForm,
				Form: &apps.Form{
					Title: "Test",
					Call: &apps.Call{
						Path: "/url",
					},
					Fields: []*apps.Field{
						{
							Type: apps.FieldTypeStaticSelect,
							Name: "field1",
							SelectStaticOptions: []apps.SelectOption{
								{
									Label: "same",
									Value: "opt1",
								},
							},
						},
					},
				},
			},
			apiExpectations: func(api *plugintest.API) {
				api.On("LogDebug", "App from malformed: Repeated label on select option. Only getting first value with the label", "field", &apps.Field{
					Type: apps.FieldTypeStaticSelect,
					Name: "field1",
					SelectStaticOptions: []apps.SelectOption{
						{
							Label: "same",
							Value: "opt1",
						},
						{
							Label: "same",
							Value: "opt2",
						},
					},
				}, "option", apps.SelectOption{
					Label: "same",
					Value: "opt2",
				}).Times(1)
			},
		},
		{
			name: "field filter static options with same value",
			appResponse: &apps.CallResponse{
				Type: apps.CallResponseTypeForm,
				Form: &apps.Form{
					Title: "Test",
					Call: &apps.Call{
						Path: "/url",
					},
					Fields: []*apps.Field{
						{
							Type: apps.FieldTypeStaticSelect,
							Name: "field1",
							SelectStaticOptions: []apps.SelectOption{
								{
									Label: "opt1",
									Value: "same",
								},
								{
									Label: "opt2",
									Value: "same",
								},
							},
						},
					},
				},
			},
			result: &apps.CallResponse{
				Type: apps.CallResponseTypeForm,
				Form: &apps.Form{
					Title: "Test",
					Call: &apps.Call{
						Path: "/url",
					},
					Fields: []*apps.Field{
						{
							Type: apps.FieldTypeStaticSelect,
							Name: "field1",
							SelectStaticOptions: []apps.SelectOption{
								{
									Label: "opt1",
									Value: "same",
								},
							},
						},
					},
				},
			},
			apiExpectations: func(api *plugintest.API) {
				api.On("LogDebug", "App from malformed: Repeated value on select option. Only getting first value with the value", "field", &apps.Field{
					Type: apps.FieldTypeStaticSelect,
					Name: "field1",
					SelectStaticOptions: []apps.SelectOption{
						{
							Label: "opt1",
							Value: "same",
						},
						{
							Label: "opt2",
							Value: "same",
						},
					},
				}, "option", apps.SelectOption{
					Label: "opt2",
					Value: "same",
				}).Times(1)
			},
		},
		{
			name: "invalid static options don't consume namespace",
			appResponse: &apps.CallResponse{
				Type: apps.CallResponseTypeForm,
				Form: &apps.Form{
					Title: "Test",
					Call: &apps.Call{
						Path: "/url",
					},
					Fields: []*apps.Field{
						{
							Type: apps.FieldTypeStaticSelect,
							Name: "field1",
							SelectStaticOptions: []apps.SelectOption{
								{
									Label: "same1",
									Value: "same1",
								},
								{
									Label: "same1",
									Value: "same2",
								},
								{
									Label: "same2",
									Value: "same1",
								},
								{
									Label: "same2",
									Value: "same2",
								},
							},
						},
					},
				},
			},
			result: &apps.CallResponse{
				Type: apps.CallResponseTypeForm,
				Form: &apps.Form{
					Title: "Test",
					Call: &apps.Call{
						Path: "/url",
					},
					Fields: []*apps.Field{
						{
							Type: apps.FieldTypeStaticSelect,
							Name: "field1",
							SelectStaticOptions: []apps.SelectOption{
								{
									Label: "same1",
									Value: "same1",
								},
								{
									Label: "same2",
									Value: "same2",
								},
							},
						},
					},
				},
			},
			apiExpectations: func(api *plugintest.API) {
				api.On("LogDebug", "App from malformed: Repeated label on select option. Only getting first value with the label", "field", &apps.Field{
					Type: apps.FieldTypeStaticSelect,
					Name: "field1",
					SelectStaticOptions: []apps.SelectOption{
						{
							Label: "same1",
							Value: "same1",
						},
						{
							Label: "same1",
							Value: "same2",
						},
						{
							Label: "same2",
							Value: "same1",
						},
						{
							Label: "same2",
							Value: "same2",
						},
					},
				}, "option", apps.SelectOption{
					Label: "same1",
					Value: "same2",
				}).Times(1)
				api.On("LogDebug", "App from malformed: Repeated value on select option. Only getting first value with the value", "field", &apps.Field{
					Type: apps.FieldTypeStaticSelect,
					Name: "field1",
					SelectStaticOptions: []apps.SelectOption{
						{
							Label: "same1",
							Value: "same1",
						},
						{
							Label: "same1",
							Value: "same2",
						},
						{
							Label: "same2",
							Value: "same1",
						},
						{
							Label: "same2",
							Value: "same2",
						},
					},
				}, "option", apps.SelectOption{
					Label: "same2",
					Value: "same1",
				}).Times(1)
			},
		},
		{
			name: "field filter static with no valid options",
			appResponse: &apps.CallResponse{
				Type: apps.CallResponseTypeForm,
				Form: &apps.Form{
					Title: "Test",
					Call: &apps.Call{
						Path: "/url",
					},
					Fields: []*apps.Field{
						{
							Type: apps.FieldTypeStaticSelect,
							Name: "field1",
							SelectStaticOptions: []apps.SelectOption{
								{},
							},
						},
					},
				},
			},
			result: &apps.CallResponse{
				Type: apps.CallResponseTypeForm,
				Form: &apps.Form{
					Title: "Test",
					Call: &apps.Call{
						Path: "/url",
					},
					Fields: []*apps.Field{},
				},
			},
			apiExpectations: func(api *plugintest.API) {
				api.On("LogDebug", "App from malformed: Option with no label", "field", &apps.Field{
					Type: apps.FieldTypeStaticSelect,
					Name: "field1",
					SelectStaticOptions: []apps.SelectOption{
						{},
					},
				}, "option value", "").Times(1)
				api.On("LogDebug", "App from malformed: Static field without opions.", "label", "field1").Times(1)
			},
		},
		{
			name: "invalid static field does not consume namespace",
			appResponse: &apps.CallResponse{
				Type: apps.CallResponseTypeForm,
				Form: &apps.Form{
					Title: "Test",
					Call: &apps.Call{
						Path: "/url",
					},
					Fields: []*apps.Field{
						{
							Type: apps.FieldTypeStaticSelect,
							Name: "field1",
							SelectStaticOptions: []apps.SelectOption{
								{},
							},
						},
						{
							Name: "field1",
						},
					},
				},
			},
			result: &apps.CallResponse{
				Type: apps.CallResponseTypeForm,
				Form: &apps.Form{
					Title: "Test",
					Call: &apps.Call{
						Path: "/url",
					},
					Fields: []*apps.Field{
						{
							Name: "field1",
						},
					},
				},
			},
			apiExpectations: func(api *plugintest.API) {
				api.On("LogDebug", "App from malformed: Option with no label", "field", &apps.Field{
					Type: apps.FieldTypeStaticSelect,
					Name: "field1",
					SelectStaticOptions: []apps.SelectOption{
						{},
					},
				}, "option value", "").Times(1)
				api.On("LogDebug", "App from malformed: Static field without opions.", "label", "field1").Times(1)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			p := newTestProxy(testApps, ctrl, tc.appResponse, tc.apiExpectations)
			c := &apps.CallRequest{
				Call: apps.Call{
					Path: "/some_path",
				},
				Context: &apps.Context{
					UserAgentContext: apps.UserAgentContext{
						AppID: "app1",
					},
				},
			}
			resp := p.Call("session_id", "acting_user_id", c)
			require.Equal(t, tc.result, resp.CallResponse)
		})
	}
}

func newTestProxy(testApps []*apps.App, ctrl *gomock.Controller, mockedResponse *apps.CallResponse, apiExpectations func(api *plugintest.API)) *Proxy {
	conf, testAPI := config.NewTestService(nil)

	testAPI.On("GetUser", mock.Anything).Return(&model.User{Locale: "en"}, nil)

	if apiExpectations != nil {
		apiExpectations(testAPI)
	}

	conf = conf.WithMattermostConfig(model.Config{
		ServiceSettings: model.ServiceSettings{
			SiteURL: model.NewString("test.mattermost.com"),
		},
	})

	s := store.NewService(conf, nil, "")
	appStore := mock_store.NewMockAppStore(ctrl)
	s.App = appStore

	upstreams := map[apps.AppID]upstream.Upstream{}
	for _, app := range testApps {
		cr := &apps.CallResponse{
			Type: apps.CallResponseTypeOK,
		}
		if mockedResponse != nil {
			cr = mockedResponse
		}
		b, _ := json.Marshal(cr)
		reader := ioutil.NopCloser(bytes.NewReader(b))

		up := mock_upstream.NewMockUpstream(ctrl)
		up.EXPECT().Roundtrip(gomock.Any(), gomock.Any()).Return(reader, nil)
		upstreams[app.Manifest.AppID] = up
		appStore.EXPECT().Get(app.AppID).Return(app, nil)
	}

	p := &Proxy{
		store:            s,
		builtinUpstreams: upstreams,
		conf:             conf,
	}

	return p
}
