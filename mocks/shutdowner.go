// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// Shutdowner is an autogenerated mock type for the Shutdowner type
type Shutdowner struct {
	mock.Mock
}

type Shutdowner_Expecter struct {
	mock *mock.Mock
}

func (_m *Shutdowner) EXPECT() *Shutdowner_Expecter {
	return &Shutdowner_Expecter{mock: &_m.Mock}
}

// Shutdown provides a mock function with given fields: ctx
func (_m *Shutdowner) Shutdown(ctx context.Context) error {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for Shutdown")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Shutdowner_Shutdown_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Shutdown'
type Shutdowner_Shutdown_Call struct {
	*mock.Call
}

// Shutdown is a helper method to define mock.On call
//   - ctx context.Context
func (_e *Shutdowner_Expecter) Shutdown(ctx interface{}) *Shutdowner_Shutdown_Call {
	return &Shutdowner_Shutdown_Call{Call: _e.mock.On("Shutdown", ctx)}
}

func (_c *Shutdowner_Shutdown_Call) Run(run func(ctx context.Context)) *Shutdowner_Shutdown_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *Shutdowner_Shutdown_Call) Return(_a0 error) *Shutdowner_Shutdown_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Shutdowner_Shutdown_Call) RunAndReturn(run func(context.Context) error) *Shutdowner_Shutdown_Call {
	_c.Call.Return(run)
	return _c
}

// NewShutdowner creates a new instance of Shutdowner. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewShutdowner(t interface {
	mock.TestingT
	Cleanup(func())
}) *Shutdowner {
	mock := &Shutdowner{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
