package gworker

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

type Worker interface {
	IsBusy() bool
	Push(job Job) error
	Run()
	Stop()
	GetTotalJob() int
	GetWorkerId() int
	Status() uint
	SetErrorHandle(ErrorHandle)
	SetJobRunOverHandle(JobRunOverHandle)
	PreSecondDealNum(num int)
}

type WorkerPool interface {
	Push(job Job) error
	Run()
	Stop()
	RecycleWorker(worker Worker)
	Status() uint
	SetErrorHandle(ErrorHandle)
	GetErrorHandle() ErrorHandle
	SetJobRunOverHandle(JobRunOverHandle)
	GetJobRunOverHandle() JobRunOverHandle
	PreSecondDealNum(num int)
	IsStop() bool
}
