// Code generated by mockery v2.35.2. DO NOT EDIT.

package mocks

import (
	rmq "github.com/aliforever/go-rmq"
	mock "github.com/stretchr/testify/mock"
)

// ConsumerBuilderImpl is an autogenerated mock type for the ConsumerBuilderImpl type
type ConsumerBuilderImpl struct {
	mock.Mock
}

// AddArg provides a mock function with given fields: key, val
func (_m *ConsumerBuilderImpl) AddArg(key string, val interface{}) rmq.ConsumerBuilderImpl {
	ret := _m.Called(key, val)

	var r0 rmq.ConsumerBuilderImpl
	if rf, ok := ret.Get(0).(func(string, interface{}) rmq.ConsumerBuilderImpl); ok {
		r0 = rf(key, val)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(rmq.ConsumerBuilderImpl)
		}
	}

	return r0
}

// Build provides a mock function with given fields:
func (_m *ConsumerBuilderImpl) Build() (rmq.ConsumerImpl, error) {
	ret := _m.Called()

	var r0 rmq.ConsumerImpl
	var r1 error
	if rf, ok := ret.Get(0).(func() (rmq.ConsumerImpl, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() rmq.ConsumerImpl); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(rmq.ConsumerImpl)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SetAutoAck provides a mock function with given fields:
func (_m *ConsumerBuilderImpl) SetAutoAck() rmq.ConsumerBuilderImpl {
	ret := _m.Called()

	var r0 rmq.ConsumerBuilderImpl
	if rf, ok := ret.Get(0).(func() rmq.ConsumerBuilderImpl); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(rmq.ConsumerBuilderImpl)
		}
	}

	return r0
}

// SetExclusive provides a mock function with given fields:
func (_m *ConsumerBuilderImpl) SetExclusive() rmq.ConsumerBuilderImpl {
	ret := _m.Called()

	var r0 rmq.ConsumerBuilderImpl
	if rf, ok := ret.Get(0).(func() rmq.ConsumerBuilderImpl); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(rmq.ConsumerBuilderImpl)
		}
	}

	return r0
}

// SetNoLocal provides a mock function with given fields:
func (_m *ConsumerBuilderImpl) SetNoLocal() rmq.ConsumerBuilderImpl {
	ret := _m.Called()

	var r0 rmq.ConsumerBuilderImpl
	if rf, ok := ret.Get(0).(func() rmq.ConsumerBuilderImpl); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(rmq.ConsumerBuilderImpl)
		}
	}

	return r0
}

// SetNoWait provides a mock function with given fields:
func (_m *ConsumerBuilderImpl) SetNoWait() rmq.ConsumerBuilderImpl {
	ret := _m.Called()

	var r0 rmq.ConsumerBuilderImpl
	if rf, ok := ret.Get(0).(func() rmq.ConsumerBuilderImpl); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(rmq.ConsumerBuilderImpl)
		}
	}

	return r0
}

// SetPrefetch provides a mock function with given fields: prefetch
func (_m *ConsumerBuilderImpl) SetPrefetch(prefetch int) rmq.ConsumerBuilderImpl {
	ret := _m.Called(prefetch)

	var r0 rmq.ConsumerBuilderImpl
	if rf, ok := ret.Get(0).(func(int) rmq.ConsumerBuilderImpl); ok {
		r0 = rf(prefetch)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(rmq.ConsumerBuilderImpl)
		}
	}

	return r0
}

// NewConsumerBuilderImpl creates a new instance of ConsumerBuilderImpl. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewConsumerBuilderImpl(t interface {
	mock.TestingT
	Cleanup(func())
}) *ConsumerBuilderImpl {
	mock := &ConsumerBuilderImpl{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
