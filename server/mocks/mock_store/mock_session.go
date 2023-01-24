// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/mattermost/mattermost-plugin-apps/server/store (interfaces: SessionStore)

// Package mock_store is a generated GoMock package.
package mock_store

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	apps "github.com/mattermost/mattermost-plugin-apps/apps"
	incoming "github.com/mattermost/mattermost-plugin-apps/server/incoming"
	model "github.com/mattermost/mattermost-server/v6/model"
)

// MockSessionStore is a mock of SessionStore interface.
type MockSessionStore struct {
	ctrl     *gomock.Controller
	recorder *MockSessionStoreMockRecorder
}

// MockSessionStoreMockRecorder is the mock recorder for MockSessionStore.
type MockSessionStoreMockRecorder struct {
	mock *MockSessionStore
}

// NewMockSessionStore creates a new mock instance.
func NewMockSessionStore(ctrl *gomock.Controller) *MockSessionStore {
	mock := &MockSessionStore{ctrl: ctrl}
	mock.recorder = &MockSessionStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSessionStore) EXPECT() *MockSessionStoreMockRecorder {
	return m.recorder
}

// Delete mocks base method.
func (m *MockSessionStore) Delete(arg0 apps.AppID, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockSessionStoreMockRecorder) Delete(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockSessionStore)(nil).Delete), arg0, arg1)
}

// DeleteAllForApp mocks base method.
func (m *MockSessionStore) DeleteAllForApp(arg0 *incoming.Request, arg1 apps.AppID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteAllForApp", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteAllForApp indicates an expected call of DeleteAllForApp.
func (mr *MockSessionStoreMockRecorder) DeleteAllForApp(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteAllForApp", reflect.TypeOf((*MockSessionStore)(nil).DeleteAllForApp), arg0, arg1)
}

// DeleteAllForUser mocks base method.
func (m *MockSessionStore) DeleteAllForUser(arg0 *incoming.Request, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteAllForUser", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteAllForUser indicates an expected call of DeleteAllForUser.
func (mr *MockSessionStoreMockRecorder) DeleteAllForUser(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteAllForUser", reflect.TypeOf((*MockSessionStore)(nil).DeleteAllForUser), arg0, arg1)
}

// Get mocks base method.
func (m *MockSessionStore) Get(arg0 apps.AppID, arg1 string) (*model.Session, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", arg0, arg1)
	ret0, _ := ret[0].(*model.Session)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockSessionStoreMockRecorder) Get(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockSessionStore)(nil).Get), arg0, arg1)
}

// ListForApp mocks base method.
func (m *MockSessionStore) ListForApp(arg0 apps.AppID) ([]*model.Session, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListForApp", arg0)
	ret0, _ := ret[0].([]*model.Session)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListForApp indicates an expected call of ListForApp.
func (mr *MockSessionStoreMockRecorder) ListForApp(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListForApp", reflect.TypeOf((*MockSessionStore)(nil).ListForApp), arg0)
}

// ListForUser mocks base method.
func (m *MockSessionStore) ListForUser(arg0 *incoming.Request, arg1 string) ([]*model.Session, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListForUser", arg0, arg1)
	ret0, _ := ret[0].([]*model.Session)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListForUser indicates an expected call of ListForUser.
func (mr *MockSessionStoreMockRecorder) ListForUser(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListForUser", reflect.TypeOf((*MockSessionStore)(nil).ListForUser), arg0, arg1)
}

// ListUsersWithSessions mocks base method.
func (m *MockSessionStore) ListUsersWithSessions(arg0 apps.AppID) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListUsersWithSessions", arg0)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListUsersWithSessions indicates an expected call of ListUsersWithSessions.
func (mr *MockSessionStoreMockRecorder) ListUsersWithSessions(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListUsersWithSessions", reflect.TypeOf((*MockSessionStore)(nil).ListUsersWithSessions), arg0)
}

// Save mocks base method.
func (m *MockSessionStore) Save(arg0 apps.AppID, arg1 string, arg2 *model.Session) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Save", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Save indicates an expected call of Save.
func (mr *MockSessionStoreMockRecorder) Save(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Save", reflect.TypeOf((*MockSessionStore)(nil).Save), arg0, arg1, arg2)
}
