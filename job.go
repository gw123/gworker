package gworker

import "time"

// Job 任务数据
type Job interface {
	GetName() string
	RetryCount() int
	Delay() time.Duration
}

// Jobber 定义如何处理任务 ()
type Jobber interface {
	Job
	Handle() error
}

type ErrorHandle func(err error, job Job)
type JobRunOverHandle func(worker Worker, job Jobber)

