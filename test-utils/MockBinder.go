// Code generated by mockery. DO NOT EDIT.

package test_utils

import (
	context "context"

	client "github.com/dapr/go-sdk/client"

	mock "github.com/stretchr/testify/mock"
)

// MockBinder is an autogenerated mock type for the Binder type
type MockBinder struct {
	mock.Mock
}

type MockBinder_Expecter struct {
	mock *mock.Mock
}

func (_m *MockBinder) EXPECT() *MockBinder_Expecter {
	return &MockBinder_Expecter{mock: &_m.Mock}
}

// InvokeBinding provides a mock function with given fields: ctx, in
func (_m *MockBinder) InvokeBinding(ctx context.Context, in *client.InvokeBindingRequest) (*client.BindingEvent, error) {
	ret := _m.Called(ctx, in)

	var r0 *client.BindingEvent
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *client.InvokeBindingRequest) (*client.BindingEvent, error)); ok {
		return rf(ctx, in)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *client.InvokeBindingRequest) *client.BindingEvent); ok {
		r0 = rf(ctx, in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*client.BindingEvent)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *client.InvokeBindingRequest) error); ok {
		r1 = rf(ctx, in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockBinder_InvokeBinding_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'InvokeBinding'
type MockBinder_InvokeBinding_Call struct {
	*mock.Call
}

// InvokeBinding is a helper method to define mock.On call
//   - ctx context.Context
//   - in *client.InvokeBindingRequest
func (_e *MockBinder_Expecter) InvokeBinding(ctx interface{}, in interface{}) *MockBinder_InvokeBinding_Call {
	return &MockBinder_InvokeBinding_Call{Call: _e.mock.On("InvokeBinding", ctx, in)}
}

func (_c *MockBinder_InvokeBinding_Call) Run(run func(ctx context.Context, in *client.InvokeBindingRequest)) *MockBinder_InvokeBinding_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*client.InvokeBindingRequest))
	})
	return _c
}

func (_c *MockBinder_InvokeBinding_Call) Return(out *client.BindingEvent, err error) *MockBinder_InvokeBinding_Call {
	_c.Call.Return(out, err)
	return _c
}

func (_c *MockBinder_InvokeBinding_Call) RunAndReturn(run func(context.Context, *client.InvokeBindingRequest) (*client.BindingEvent, error)) *MockBinder_InvokeBinding_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockBinder creates a new instance of MockBinder. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockBinder(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockBinder {
	mock := &MockBinder{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
