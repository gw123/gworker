package gworker

import (
	"context"
)

type Job interface {
	UUID() string
	Queue() string
	Delay() int
	Marshal() ([]byte, error)
	JobHandler(ctx context.Context, job Jobber) error
}

/**
  JobHandler  operate Job
   1. Body() get job payload
   2. Attempt()  current msg tried times
   3. Ok() ack current message
   4. Skip() nack current message
   5. Retry() sends message to a deferred queue to try this message after the delay seconds
*/
type Jobber interface {
	Attempt() int
	OK() error
	//Call by JobHandler
	Skip() error
	//Call by JobHandler
	Retry(ctx context.Context, delay int) error
	Body() []byte
}

type JobHandler func(ctx context.Context, job Jobber) error

type JobManager interface {
	Dispatch(ctx context.Context, job Job) error
	Do(ctx context.Context, queue string, handler JobHandler) error
	Close() error
}
