package interfaces

import "io"

const JobFlagEnd = 2
const JobFlagNormal = 1

type Job interface {
	GetPayload() []byte
	GetCreatedTime() int64
	GetJobFlag() int64
	SetWorkerName(string2 string)
	GetWorkerName() string
	SetWriteCloser(input io.WriteCloser)
	SetReadCloser(ouput io.ReadCloser)
	DoJob()
	// 当前正在执行的任务
	Cancel()
}
