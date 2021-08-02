// Code generated by mockery v1.0.0
package mocks

import (
	context "context"
	"github.com/gw123/gworker"
)
import mock "github.com/stretchr/testify/mock"

// JobManager is an autogenerated mock type for the JobManager type
type JobManager struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *JobManager) Close() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Dispatch provides a mock function with given fields: ctx, job
func (_m *JobManager) Dispatch(ctx context.Context, job gworker.Job) error {
	ret := _m.Called(ctx, job)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, gworker.Job) error); ok {
		r0 = rf(ctx, job)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Do provides a mock function with given fields: ctx, queue, handler
func (_m *JobManager) Do(ctx context.Context, queue string, handler gworker.JobHandler) error {
	ret := _m.Called(ctx, queue, handler)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, gworker.JobHandler) error); ok {
		r0 = rf(ctx, queue, handler)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
