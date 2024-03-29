// Code generated by mockery v1.0.0
package mocks

import context "context"
import mock "github.com/stretchr/testify/mock"
import rabbiter "github.com/gw123/gworker/rabbiter"

// Publisher is an autogenerated mock type for the Publisher type
type Publisher struct {
	mock.Mock
}

// Publish provides a mock function with given fields: ctx, ex, m
func (_m *Publisher) Publish(ctx context.Context, ex *rabbiter.Exchange, m *rabbiter.Message) error {
	ret := _m.Called(ctx, ex, m)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *rabbiter.Exchange, *rabbiter.Message) error); ok {
		r0 = rf(ctx, ex, m)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// PublishWithDelay provides a mock function with given fields: ctx, q, m, delay
func (_m *Publisher) PublishWithDelay(ctx context.Context, q *rabbiter.Queue, m *rabbiter.Message, delay int) error {
	ret := _m.Called(ctx, q, m, delay)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *rabbiter.Queue, *rabbiter.Message, int) error); ok {
		r0 = rf(ctx, q, m, delay)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
