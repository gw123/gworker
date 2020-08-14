package gworker

import "io"

const (
	Worker_Running = 1
	Worker_Free    = 0
)

type ErrorHandle func(err error, job Job)
type JobRunOverHandle func(worker Worker, job Job)

type Job interface {
	GetJobType() string
	Run() error
	Stop()
}

type BaseJob struct {
	WorkerName  string
	CreatedTime int64
	UpdatedTime int64
	Flag        int64
	JobType     string
	Payload     []byte
	Response    []byte
	Input       io.WriteCloser
	Output      io.ReadCloser
	runFlag     bool
}

