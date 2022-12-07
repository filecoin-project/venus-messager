// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/filecoin-project/venus-auth/jwtclient (interfaces: IAuthClient)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	auth "github.com/filecoin-project/venus-auth/auth"
	gomock "github.com/golang/mock/gomock"
)

// MockIAuthClient is a mock of IAuthClient interface.
type MockIAuthClient struct {
	ctrl     *gomock.Controller
	recorder *MockIAuthClientMockRecorder
}

// MockIAuthClientMockRecorder is the mock recorder for MockIAuthClient.
type MockIAuthClientMockRecorder struct {
	mock *MockIAuthClient
}

// NewMockIAuthClient creates a new mock instance.
func NewMockIAuthClient(ctrl *gomock.Controller) *MockIAuthClient {
	mock := &MockIAuthClient{ctrl: ctrl}
	mock.recorder = &MockIAuthClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIAuthClient) EXPECT() *MockIAuthClientMockRecorder {
	return m.recorder
}

// GetUser mocks base method.
func (m *MockIAuthClient) GetUser(arg0 *auth.GetUserRequest) (*auth.OutputUser, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUser", arg0)
	ret0, _ := ret[0].(*auth.OutputUser)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUser indicates an expected call of GetUser.
func (mr *MockIAuthClientMockRecorder) GetUser(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUser", reflect.TypeOf((*MockIAuthClient)(nil).GetUser), arg0)
}

// GetUserByMiner mocks base method.
func (m *MockIAuthClient) GetUserByMiner(arg0 *auth.GetUserByMinerRequest) (*auth.OutputUser, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserByMiner", arg0)
	ret0, _ := ret[0].(*auth.OutputUser)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserByMiner indicates an expected call of GetUserByMiner.
func (mr *MockIAuthClientMockRecorder) GetUserByMiner(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserByMiner", reflect.TypeOf((*MockIAuthClient)(nil).GetUserByMiner), arg0)
}

// GetUserBySigner mocks base method.
func (m *MockIAuthClient) GetUserBySigner(arg0 string) ([]*auth.OutputUser, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserBySigner", arg0)
	ret0, _ := ret[0].([]*auth.OutputUser)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserBySigner indicates an expected call of GetUserBySigner.
func (mr *MockIAuthClientMockRecorder) GetUserBySigner(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserBySigner", reflect.TypeOf((*MockIAuthClient)(nil).GetUserBySigner), arg0)
}

// GetUserRateLimit mocks base method.
func (m *MockIAuthClient) GetUserRateLimit(arg0, arg1 string) (auth.GetUserRateLimitResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserRateLimit", arg0, arg1)
	ret0, _ := ret[0].(auth.GetUserRateLimitResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserRateLimit indicates an expected call of GetUserRateLimit.
func (mr *MockIAuthClientMockRecorder) GetUserRateLimit(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserRateLimit", reflect.TypeOf((*MockIAuthClient)(nil).GetUserRateLimit), arg0, arg1)
}

// HasMiner mocks base method.
func (m *MockIAuthClient) HasMiner(arg0 *auth.HasMinerRequest) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HasMiner", arg0)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// HasMiner indicates an expected call of HasMiner.
func (mr *MockIAuthClientMockRecorder) HasMiner(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HasMiner", reflect.TypeOf((*MockIAuthClient)(nil).HasMiner), arg0)
}

// HasSigner mocks base method.
func (m *MockIAuthClient) HasSigner(arg0 string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HasSigner", arg0)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// HasSigner indicates an expected call of HasSigner.
func (mr *MockIAuthClientMockRecorder) HasSigner(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HasSigner", reflect.TypeOf((*MockIAuthClient)(nil).HasSigner), arg0)
}

// HasUser mocks base method.
func (m *MockIAuthClient) HasUser(arg0 *auth.HasUserRequest) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HasUser", arg0)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// HasUser indicates an expected call of HasUser.
func (mr *MockIAuthClientMockRecorder) HasUser(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HasUser", reflect.TypeOf((*MockIAuthClient)(nil).HasUser), arg0)
}

// ListMiners mocks base method.
func (m *MockIAuthClient) ListMiners(arg0 string) (auth.ListMinerResp, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListMiners", arg0)
	ret0, _ := ret[0].(auth.ListMinerResp)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListMiners indicates an expected call of ListMiners.
func (mr *MockIAuthClientMockRecorder) ListMiners(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListMiners", reflect.TypeOf((*MockIAuthClient)(nil).ListMiners), arg0)
}

// ListSigners mocks base method.
func (m *MockIAuthClient) ListSigners(arg0 string) (auth.ListSignerResp, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListSigners", arg0)
	ret0, _ := ret[0].(auth.ListSignerResp)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListSigners indicates an expected call of ListSigners.
func (mr *MockIAuthClientMockRecorder) ListSigners(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListSigners", reflect.TypeOf((*MockIAuthClient)(nil).ListSigners), arg0)
}

// ListUsers mocks base method.
func (m *MockIAuthClient) ListUsers(arg0 *auth.ListUsersRequest) ([]*auth.OutputUser, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListUsers", arg0)
	ret0, _ := ret[0].([]*auth.OutputUser)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListUsers indicates an expected call of ListUsers.
func (mr *MockIAuthClientMockRecorder) ListUsers(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListUsers", reflect.TypeOf((*MockIAuthClient)(nil).ListUsers), arg0)
}

// ListUsersWithMiners mocks base method.
func (m *MockIAuthClient) ListUsersWithMiners(arg0 *auth.ListUsersRequest) ([]*auth.OutputUser, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListUsersWithMiners", arg0)
	ret0, _ := ret[0].([]*auth.OutputUser)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListUsersWithMiners indicates an expected call of ListUsersWithMiners.
func (mr *MockIAuthClientMockRecorder) ListUsersWithMiners(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListUsersWithMiners", reflect.TypeOf((*MockIAuthClient)(nil).ListUsersWithMiners), arg0)
}

// MinerExistInUser mocks base method.
func (m *MockIAuthClient) MinerExistInUser(arg0, arg1 string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MinerExistInUser", arg0, arg1)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MinerExistInUser indicates an expected call of MinerExistInUser.
func (mr *MockIAuthClientMockRecorder) MinerExistInUser(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MinerExistInUser", reflect.TypeOf((*MockIAuthClient)(nil).MinerExistInUser), arg0, arg1)
}

// RegisterSigners mocks base method.
func (m *MockIAuthClient) RegisterSigners(arg0 string, arg1 []string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RegisterSigners", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// RegisterSigners indicates an expected call of RegisterSigners.
func (mr *MockIAuthClientMockRecorder) RegisterSigners(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterSigners", reflect.TypeOf((*MockIAuthClient)(nil).RegisterSigners), arg0, arg1)
}

// SignerExistInUser mocks base method.
func (m *MockIAuthClient) SignerExistInUser(arg0, arg1 string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SignerExistInUser", arg0, arg1)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SignerExistInUser indicates an expected call of SignerExistInUser.
func (mr *MockIAuthClientMockRecorder) SignerExistInUser(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SignerExistInUser", reflect.TypeOf((*MockIAuthClient)(nil).SignerExistInUser), arg0, arg1)
}

// UnregisterSigners mocks base method.
func (m *MockIAuthClient) UnregisterSigners(arg0 string, arg1 []string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UnregisterSigners", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UnregisterSigners indicates an expected call of UnregisterSigners.
func (mr *MockIAuthClientMockRecorder) UnregisterSigners(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UnregisterSigners", reflect.TypeOf((*MockIAuthClient)(nil).UnregisterSigners), arg0, arg1)
}

// VerifyUsers mocks base method.
func (m *MockIAuthClient) VerifyUsers(arg0 []string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "VerifyUsers", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// VerifyUsers indicates an expected call of VerifyUsers.
func (mr *MockIAuthClientMockRecorder) VerifyUsers(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "VerifyUsers", reflect.TypeOf((*MockIAuthClient)(nil).VerifyUsers), arg0)
}
