// Code generated by MockGen. DO NOT EDIT.
// Source: application/service/git/git.go

// Package mock_git is a generated GoMock package.
package mock_git

import (
	context "github.com/duck8823/duci/application/context"
	gomock "github.com/golang/mock/gomock"
	plumbing "gopkg.in/src-d/go-git.v4/plumbing"
	reflect "reflect"
)

// MockService is a mock of Service interface
type MockService struct {
	ctrl     *gomock.Controller
	recorder *MockServiceMockRecorder
}

// MockServiceMockRecorder is the mock recorder for MockService
type MockServiceMockRecorder struct {
	mock *MockService
}

// NewMockService creates a new mock instance
func NewMockService(ctrl *gomock.Controller) *MockService {
	mock := &MockService{ctrl: ctrl}
	mock.recorder = &MockServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockService) EXPECT() *MockServiceMockRecorder {
	return m.recorder
}

// Clone mocks base method
func (m *MockService) Clone(ctx context.Context, dir, sshUrl, ref string, sha plumbing.Hash) error {
	ret := m.ctrl.Call(m, "Clone", ctx, dir, sshUrl, ref, sha)
	ret0, _ := ret[0].(error)
	return ret0
}

// Clone indicates an expected call of Clone
func (mr *MockServiceMockRecorder) Clone(ctx, dir, sshUrl, ref, sha interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Clone", reflect.TypeOf((*MockService)(nil).Clone), ctx, dir, sshUrl, ref, sha)
}
