// Code generated by mockery v2.35.2. DO NOT EDIT.

package mocks

import (
	rmq "github.com/aliforever/go-rmq"
	mock "github.com/stretchr/testify/mock"

	time "time"
)

// PublishFieldsImpl is an autogenerated mock type for the PublishFieldsImpl type
type PublishFieldsImpl struct {
	mock.Mock
}

// AddHeader provides a mock function with given fields: key, val
func (_m *PublishFieldsImpl) AddHeader(key string, val interface{}) rmq.PublisherBuilderImpl {
	ret := _m.Called(key, val)

	var r0 rmq.PublisherBuilderImpl
	if rf, ok := ret.Get(0).(func(string, interface{}) rmq.PublisherBuilderImpl); ok {
		r0 = rf(key, val)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(rmq.PublisherBuilderImpl)
		}
	}

	return r0
}

// DeliveryModePersistent provides a mock function with given fields:
func (_m *PublishFieldsImpl) DeliveryModePersistent() rmq.PublisherBuilderImpl {
	ret := _m.Called()

	var r0 rmq.PublisherBuilderImpl
	if rf, ok := ret.Get(0).(func() rmq.PublisherBuilderImpl); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(rmq.PublisherBuilderImpl)
		}
	}

	return r0
}

// DeliveryModeTransient provides a mock function with given fields:
func (_m *PublishFieldsImpl) DeliveryModeTransient() rmq.PublisherBuilderImpl {
	ret := _m.Called()

	var r0 rmq.PublisherBuilderImpl
	if rf, ok := ret.Get(0).(func() rmq.PublisherBuilderImpl); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(rmq.PublisherBuilderImpl)
		}
	}

	return r0
}

// SetContentType provides a mock function with given fields: contentType
func (_m *PublishFieldsImpl) SetContentType(contentType string) rmq.PublisherBuilderImpl {
	ret := _m.Called(contentType)

	var r0 rmq.PublisherBuilderImpl
	if rf, ok := ret.Get(0).(func(string) rmq.PublisherBuilderImpl); ok {
		r0 = rf(contentType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(rmq.PublisherBuilderImpl)
		}
	}

	return r0
}

// SetCorrelationID provides a mock function with given fields: id
func (_m *PublishFieldsImpl) SetCorrelationID(id string) rmq.PublisherBuilderImpl {
	ret := _m.Called(id)

	var r0 rmq.PublisherBuilderImpl
	if rf, ok := ret.Get(0).(func(string) rmq.PublisherBuilderImpl); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(rmq.PublisherBuilderImpl)
		}
	}

	return r0
}

// SetDataTypeBytes provides a mock function with given fields:
func (_m *PublishFieldsImpl) SetDataTypeBytes() rmq.PublisherBuilderImpl {
	ret := _m.Called()

	var r0 rmq.PublisherBuilderImpl
	if rf, ok := ret.Get(0).(func() rmq.PublisherBuilderImpl); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(rmq.PublisherBuilderImpl)
		}
	}

	return r0
}

// SetDataTypeJSON provides a mock function with given fields:
func (_m *PublishFieldsImpl) SetDataTypeJSON() rmq.PublisherBuilderImpl {
	ret := _m.Called()

	var r0 rmq.PublisherBuilderImpl
	if rf, ok := ret.Get(0).(func() rmq.PublisherBuilderImpl); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(rmq.PublisherBuilderImpl)
		}
	}

	return r0
}

// SetExpiration provides a mock function with given fields: dur
func (_m *PublishFieldsImpl) SetExpiration(dur time.Duration) rmq.PublisherBuilderImpl {
	ret := _m.Called(dur)

	var r0 rmq.PublisherBuilderImpl
	if rf, ok := ret.Get(0).(func(time.Duration) rmq.PublisherBuilderImpl); ok {
		r0 = rf(dur)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(rmq.PublisherBuilderImpl)
		}
	}

	return r0
}

// SetImmediate provides a mock function with given fields:
func (_m *PublishFieldsImpl) SetImmediate() rmq.PublisherBuilderImpl {
	ret := _m.Called()

	var r0 rmq.PublisherBuilderImpl
	if rf, ok := ret.Get(0).(func() rmq.PublisherBuilderImpl); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(rmq.PublisherBuilderImpl)
		}
	}

	return r0
}

// SetMandatory provides a mock function with given fields:
func (_m *PublishFieldsImpl) SetMandatory() rmq.PublisherBuilderImpl {
	ret := _m.Called()

	var r0 rmq.PublisherBuilderImpl
	if rf, ok := ret.Get(0).(func() rmq.PublisherBuilderImpl); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(rmq.PublisherBuilderImpl)
		}
	}

	return r0
}

// SetReplyToID provides a mock function with given fields: id
func (_m *PublishFieldsImpl) SetReplyToID(id string) rmq.PublisherBuilderImpl {
	ret := _m.Called(id)

	var r0 rmq.PublisherBuilderImpl
	if rf, ok := ret.Get(0).(func(string) rmq.PublisherBuilderImpl); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(rmq.PublisherBuilderImpl)
		}
	}

	return r0
}

// NewPublishFieldsImpl creates a new instance of PublishFieldsImpl. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewPublishFieldsImpl(t interface {
	mock.TestingT
	Cleanup(func())
}) *PublishFieldsImpl {
	mock := &PublishFieldsImpl{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
