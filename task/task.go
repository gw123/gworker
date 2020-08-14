package task

import (
	"encoding/json"
)

//只关心任务数据
type Task interface {
	TaskName() string
	RequestID() string
	Trace() []string
	json.Marshaler
	json.Unmarshaler
}

//如何处理任务 Task
type TaskHandle interface {
	GetTaskName() string
	HandleFun(data []byte) error
}


