// Code generated by mockery v1.0.0
package mocks

import mock "github.com/stretchr/testify/mock"
import rabbiter "github.com/gw123/gworker/rabbiter"

// Factory is an autogenerated mock type for the Factory type
type Factory struct {
	mock.Mock
}

// Consumer provides a mock function with given fields: _a0, _a1, _a2
func (_m *Factory) Consumer(_a0 *rabbiter.Exchange, _a1 *rabbiter.Queue, _a2 rabbiter.HandlerFunc) rabbiter.Consumer {
	ret := _m.Called(_a0, _a1, _a2)

	var r0 rabbiter.Consumer
	if rf, ok := ret.Get(0).(func(*rabbiter.Exchange, *rabbiter.Queue, rabbiter.HandlerFunc) rabbiter.Consumer); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(rabbiter.Consumer)
		}
	}

	return r0
}

// Publisher provides a mock function with given fields:
func (_m *Factory) Publisher() rabbiter.Publisher {
	ret := _m.Called()

	var r0 rabbiter.Publisher
	if rf, ok := ret.Get(0).(func() rabbiter.Publisher); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(rabbiter.Publisher)
		}
	}

	return r0
}