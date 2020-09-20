package gworker

import "time"

// Task 任务数据
type Task interface {
	GetName() string
	GetID() string
	Trace() []string
	RetryCount() int
	Delay() time.Duration
}

//eta := time.Now().UTC().Add(time.Second * 5)
//signature.ETA = &eta

// TaskHandle 定义如何处理任务 ()
type TaskHandle interface {
	GetName() string
	HandleFun(data string) error
}

// Tasker 相当于 Task和taskHandle结合，包含数据和执行任务方法
type Tasker interface {
	Task
	Handle() error
}

