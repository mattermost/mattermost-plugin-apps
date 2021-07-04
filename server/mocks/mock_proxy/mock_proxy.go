// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/mattermost/mattermost-plugin-apps/server/proxy (interfaces: Service)

// Package mock_proxy is a generated GoMock package.
package mock_proxy

import (
	io "io"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	apps "github.com/mattermost/mattermost-plugin-apps/apps"
	mmclient "github.com/mattermost/mattermost-plugin-apps/mmclient"
	upstream "github.com/mattermost/mattermost-plugin-apps/upstream"
	md "github.com/mattermost/mattermost-plugin-apps/utils/md"
)

// MockService is a mock of Service interface.
type MockService struct {
	ctrl     *gomock.Controller
	recorder *MockServiceMockRecorder
}

// MockServiceMockRecorder is the mock recorder for MockService.
type MockServiceMockRecorder struct {
	mock *MockService
}

// NewMockService creates a new mock instance.
func NewMockService(ctrl *gomock.Controller) *MockService {
	mock := &MockService{ctrl: ctrl}
	mock.recorder = &MockServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockService) EXPECT() *MockServiceMockRecorder {
	return m.recorder
}

// AddBuiltinUpstream mocks base method.
func (m *MockService) AddBuiltinUpstream(arg0 apps.AppID, arg1 upstream.Upstream) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddBuiltinUpstream", arg0, arg1)
}

// AddBuiltinUpstream indicates an expected call of AddBuiltinUpstream.
func (mr *MockServiceMockRecorder) AddBuiltinUpstream(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddBuiltinUpstream", reflect.TypeOf((*MockService)(nil).AddBuiltinUpstream), arg0, arg1)
}

// AddLocalManifest mocks base method.
func (m *MockService) AddLocalManifest(arg0 string, arg1 *apps.Manifest) (md.MD, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddLocalManifest", arg0, arg1)
	ret0, _ := ret[0].(md.MD)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AddLocalManifest indicates an expected call of AddLocalManifest.
func (mr *MockServiceMockRecorder) AddLocalManifest(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddLocalManifest", reflect.TypeOf((*MockService)(nil).AddLocalManifest), arg0, arg1)
}

// AppIsEnabled mocks base method.
func (m *MockService) AppIsEnabled(arg0 *apps.App) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AppIsEnabled", arg0)
	ret0, _ := ret[0].(bool)
	return ret0
}

// AppIsEnabled indicates an expected call of AppIsEnabled.
func (mr *MockServiceMockRecorder) AppIsEnabled(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AppIsEnabled", reflect.TypeOf((*MockService)(nil).AppIsEnabled), arg0)
}

// Call mocks base method.
func (m *MockService) Call(arg0, arg1 string, arg2 *apps.CallRequest) *apps.ProxyCallResponse {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Call", arg0, arg1, arg2)
	ret0, _ := ret[0].(*apps.ProxyCallResponse)
	return ret0
}

// Call indicates an expected call of Call.
func (mr *MockServiceMockRecorder) Call(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Call", reflect.TypeOf((*MockService)(nil).Call), arg0, arg1, arg2)
}

// CompleteRemoteOAuth2 mocks base method.
func (m *MockService) CompleteRemoteOAuth2(arg0, arg1 string, arg2 apps.AppID, arg3 map[string]interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CompleteRemoteOAuth2", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// CompleteRemoteOAuth2 indicates an expected call of CompleteRemoteOAuth2.
func (mr *MockServiceMockRecorder) CompleteRemoteOAuth2(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CompleteRemoteOAuth2", reflect.TypeOf((*MockService)(nil).CompleteRemoteOAuth2), arg0, arg1, arg2, arg3)
}

// DisableApp mocks base method.
func (m *MockService) DisableApp(arg0 mmclient.Client, arg1 string, arg2 *apps.Context, arg3 apps.AppID) (md.MD, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DisableApp", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(md.MD)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DisableApp indicates an expected call of DisableApp.
func (mr *MockServiceMockRecorder) DisableApp(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DisableApp", reflect.TypeOf((*MockService)(nil).DisableApp), arg0, arg1, arg2, arg3)
}

// EnableApp mocks base method.
func (m *MockService) EnableApp(arg0 mmclient.Client, arg1 string, arg2 *apps.Context, arg3 apps.AppID) (md.MD, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnableApp", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(md.MD)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EnableApp indicates an expected call of EnableApp.
func (mr *MockServiceMockRecorder) EnableApp(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnableApp", reflect.TypeOf((*MockService)(nil).EnableApp), arg0, arg1, arg2, arg3)
}

// GetAsset mocks base method.
func (m *MockService) GetAsset(arg0 apps.AppID, arg1 string) (io.ReadCloser, int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAsset", arg0, arg1)
	ret0, _ := ret[0].(io.ReadCloser)
	ret1, _ := ret[1].(int)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetAsset indicates an expected call of GetAsset.
func (mr *MockServiceMockRecorder) GetAsset(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAsset", reflect.TypeOf((*MockService)(nil).GetAsset), arg0, arg1)
}

// GetBindings mocks base method.
func (m *MockService) GetBindings(arg0, arg1 string, arg2 *apps.Context) ([]*apps.Binding, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBindings", arg0, arg1, arg2)
	ret0, _ := ret[0].([]*apps.Binding)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBindings indicates an expected call of GetBindings.
func (mr *MockServiceMockRecorder) GetBindings(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBindings", reflect.TypeOf((*MockService)(nil).GetBindings), arg0, arg1, arg2)
}

// GetInstalledApp mocks base method.
func (m *MockService) GetInstalledApp(arg0 apps.AppID) (*apps.App, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetInstalledApp", arg0)
	ret0, _ := ret[0].(*apps.App)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetInstalledApp indicates an expected call of GetInstalledApp.
func (mr *MockServiceMockRecorder) GetInstalledApp(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetInstalledApp", reflect.TypeOf((*MockService)(nil).GetInstalledApp), arg0)
}

// GetInstalledApps mocks base method.
func (m *MockService) GetInstalledApps() []*apps.App {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetInstalledApps")
	ret0, _ := ret[0].([]*apps.App)
	return ret0
}

// GetInstalledApps indicates an expected call of GetInstalledApps.
func (mr *MockServiceMockRecorder) GetInstalledApps() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetInstalledApps", reflect.TypeOf((*MockService)(nil).GetInstalledApps))
}

// GetListedApps mocks base method.
func (m *MockService) GetListedApps(arg0 string, arg1 bool) []*apps.ListedApp {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetListedApps", arg0, arg1)
	ret0, _ := ret[0].([]*apps.ListedApp)
	return ret0
}

// GetListedApps indicates an expected call of GetListedApps.
func (mr *MockServiceMockRecorder) GetListedApps(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetListedApps", reflect.TypeOf((*MockService)(nil).GetListedApps), arg0, arg1)
}

// GetManifest mocks base method.
func (m *MockService) GetManifest(arg0 apps.AppID) (*apps.Manifest, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetManifest", arg0)
	ret0, _ := ret[0].(*apps.Manifest)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetManifest indicates an expected call of GetManifest.
func (mr *MockServiceMockRecorder) GetManifest(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetManifest", reflect.TypeOf((*MockService)(nil).GetManifest), arg0)
}

// GetManifestFromS3 mocks base method.
func (m *MockService) GetManifestFromS3(arg0 apps.AppID, arg1 apps.AppVersion) (*apps.Manifest, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetManifestFromS3", arg0, arg1)
	ret0, _ := ret[0].(*apps.Manifest)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetManifestFromS3 indicates an expected call of GetManifestFromS3.
func (mr *MockServiceMockRecorder) GetManifestFromS3(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetManifestFromS3", reflect.TypeOf((*MockService)(nil).GetManifestFromS3), arg0, arg1)
}

// GetRemoteOAuth2ConnectURL mocks base method.
func (m *MockService) GetRemoteOAuth2ConnectURL(arg0, arg1 string, arg2 apps.AppID) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRemoteOAuth2ConnectURL", arg0, arg1, arg2)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRemoteOAuth2ConnectURL indicates an expected call of GetRemoteOAuth2ConnectURL.
func (mr *MockServiceMockRecorder) GetRemoteOAuth2ConnectURL(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRemoteOAuth2ConnectURL", reflect.TypeOf((*MockService)(nil).GetRemoteOAuth2ConnectURL), arg0, arg1, arg2)
}

// InstallApp mocks base method.
func (m *MockService) InstallApp(arg0 mmclient.Client, arg1 string, arg2 *apps.Context, arg3 bool, arg4, arg5 string) (*apps.App, md.MD, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InstallApp", arg0, arg1, arg2, arg3, arg4, arg5)
	ret0, _ := ret[0].(*apps.App)
	ret1, _ := ret[1].(md.MD)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// InstallApp indicates an expected call of InstallApp.
func (mr *MockServiceMockRecorder) InstallApp(arg0, arg1, arg2, arg3, arg4, arg5 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InstallApp", reflect.TypeOf((*MockService)(nil).InstallApp), arg0, arg1, arg2, arg3, arg4, arg5)
}

// Notify mocks base method.
func (m *MockService) Notify(arg0 *apps.Context, arg1 apps.Subject) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Notify", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Notify indicates an expected call of Notify.
func (mr *MockServiceMockRecorder) Notify(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Notify", reflect.TypeOf((*MockService)(nil).Notify), arg0, arg1)
}

// NotifyRemoteWebhook mocks base method.
func (m *MockService) NotifyRemoteWebhook(arg0 *apps.App, arg1 []byte, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NotifyRemoteWebhook", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// NotifyRemoteWebhook indicates an expected call of NotifyRemoteWebhook.
func (mr *MockServiceMockRecorder) NotifyRemoteWebhook(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NotifyRemoteWebhook", reflect.TypeOf((*MockService)(nil).NotifyRemoteWebhook), arg0, arg1, arg2)
}

// SynchronizeInstalledApps mocks base method.
func (m *MockService) SynchronizeInstalledApps() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SynchronizeInstalledApps")
	ret0, _ := ret[0].(error)
	return ret0
}

// SynchronizeInstalledApps indicates an expected call of SynchronizeInstalledApps.
func (mr *MockServiceMockRecorder) SynchronizeInstalledApps() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SynchronizeInstalledApps", reflect.TypeOf((*MockService)(nil).SynchronizeInstalledApps))
}

// UninstallApp mocks base method.
func (m *MockService) UninstallApp(arg0 mmclient.Client, arg1 string, arg2 *apps.Context, arg3 apps.AppID) (md.MD, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UninstallApp", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(md.MD)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UninstallApp indicates an expected call of UninstallApp.
func (mr *MockServiceMockRecorder) UninstallApp(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UninstallApp", reflect.TypeOf((*MockService)(nil).UninstallApp), arg0, arg1, arg2, arg3)
}
