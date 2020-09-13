package gworker

// Task 任务数据
type Task interface {
	GetName() string
	GetID() string
	Trace() []string
}

// TaskHandle 定义如何处理任务 ()
type TaskHandle interface {
	GetName() string
	HandleFun(data string) error
}

// Tasker 相当于 Task和taskHandle结合，包含数据和执行任务方法
type Tasker interface {
	Handle() error
}


type ErrorHandle func(err error, job Tasker)
type JobRunOverHandle func(worker Worker, job Tasker)

