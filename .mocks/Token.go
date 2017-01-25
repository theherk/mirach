package mocks

import (
	mock "github.com/stretchr/testify/mock"

	time "time"
)

// Generated do not edit

// Token is an autogenerated mock type for the Token type
type Token struct {
	mock.Mock
}

// flowComplete provides a mock function with given fields:
func (_m *Token) flowComplete() {
	_m.Called()
}

// Error provides a mock function with given fields:
func (_m *Token) Error() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Wait provides a mock function with given fields:
func (_m *Token) Wait() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// WaitTimeout provides a mock function with given fields: _a0
func (_m *Token) WaitTimeout(_a0 time.Duration) bool {
	ret := _m.Called(_a0)

	var r0 bool
	if rf, ok := ret.Get(0).(func(time.Duration) bool); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// var _ mqtt.Token = (*Token)(nil)
