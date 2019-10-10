package gworker

const (
	Worker_Running = 1
	Worker_Free    = 0
)

type ErrorHandle func(err error, job Job)

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
	GetErrorHandle() ErrorHandle
	SetErrorHandle(ErrorHandle)
	PreSecondDealNum(num int)
}

type WorkerPool interface {
	Push(job Job) error
	Run()
	Stop() chan int
	RecycleWorker(worker Worker)
	Status() uint
	GetErrorHandle() ErrorHandle
	SetErrorHandle(ErrorHandle)
	PreSecondDealNum(num int)
}
