// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/mattermost/mattermost-plugin-apps/server/appservices (interfaces: Service)

// Package mock_appservices is a generated GoMock package.
package mock_appservices

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	apps "github.com/mattermost/mattermost-plugin-apps/apps"
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

// GetOAuth2User mocks base method.
func (m *MockService) GetOAuth2User(arg0 apps.AppID, arg1 string, arg2 interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOAuth2User", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// GetOAuth2User indicates an expected call of GetOAuth2User.
func (mr *MockServiceMockRecorder) GetOAuth2User(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOAuth2User", reflect.TypeOf((*MockService)(nil).GetOAuth2User), arg0, arg1, arg2)
}

// GetSubscriptions mocks base method.
func (m *MockService) GetSubscriptions(arg0 string) ([]apps.Subscription, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSubscriptions", arg0)
	ret0, _ := ret[0].([]apps.Subscription)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSubscriptions indicates an expected call of GetSubscriptions.
func (mr *MockServiceMockRecorder) GetSubscriptions(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSubscriptions", reflect.TypeOf((*MockService)(nil).GetSubscriptions), arg0)
}

// KVDelete mocks base method.
func (m *MockService) KVDelete(arg0, arg1, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "KVDelete", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// KVDelete indicates an expected call of KVDelete.
func (mr *MockServiceMockRecorder) KVDelete(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "KVDelete", reflect.TypeOf((*MockService)(nil).KVDelete), arg0, arg1, arg2)
}

// KVGet mocks base method.
func (m *MockService) KVGet(arg0, arg1, arg2 string, arg3 interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "KVGet", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// KVGet indicates an expected call of KVGet.
func (mr *MockServiceMockRecorder) KVGet(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "KVGet", reflect.TypeOf((*MockService)(nil).KVGet), arg0, arg1, arg2, arg3)
}

// KVSet mocks base method.
func (m *MockService) KVSet(arg0, arg1, arg2 string, arg3 interface{}) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "KVSet", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// KVSet indicates an expected call of KVSet.
func (mr *MockServiceMockRecorder) KVSet(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "KVSet", reflect.TypeOf((*MockService)(nil).KVSet), arg0, arg1, arg2, arg3)
}

// StoreOAuth2App mocks base method.
func (m *MockService) StoreOAuth2App(arg0 apps.AppID, arg1 string, arg2 apps.OAuth2App) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StoreOAuth2App", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// StoreOAuth2App indicates an expected call of StoreOAuth2App.
func (mr *MockServiceMockRecorder) StoreOAuth2App(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StoreOAuth2App", reflect.TypeOf((*MockService)(nil).StoreOAuth2App), arg0, arg1, arg2)
}

// StoreOAuth2User mocks base method.
func (m *MockService) StoreOAuth2User(arg0 apps.AppID, arg1 string, arg2 []byte) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StoreOAuth2User", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// StoreOAuth2User indicates an expected call of StoreOAuth2User.
func (mr *MockServiceMockRecorder) StoreOAuth2User(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StoreOAuth2User", reflect.TypeOf((*MockService)(nil).StoreOAuth2User), arg0, arg1, arg2)
}

// Subscribe mocks base method.
func (m *MockService) Subscribe(arg0 apps.Subscription) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Subscribe", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Subscribe indicates an expected call of Subscribe.
func (mr *MockServiceMockRecorder) Subscribe(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Subscribe", reflect.TypeOf((*MockService)(nil).Subscribe), arg0)
}

// Unsubscribe mocks base method.
func (m *MockService) Unsubscribe(arg0 apps.Subscription) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Unsubscribe", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Unsubscribe indicates an expected call of Unsubscribe.
func (mr *MockServiceMockRecorder) Unsubscribe(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Unsubscribe", reflect.TypeOf((*MockService)(nil).Unsubscribe), arg0)
}
