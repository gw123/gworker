package gworker

import "time"

// Job 任务数据
type Job interface {
	GetName() string
	GetID() string
	Trace() []string
	RetryCount() int
	Delay() time.Duration
}

// JobHandle 定义如何处理任务 ()
type JobHandle interface {
	GetName() string
	HandleFun(data string) error
}

type Jobber interface {
	Task
	Handle() error
}

type ErrorHandle func(err error, job Tasker)
type JobRunOverHandle func(worker Worker, job Tasker)

