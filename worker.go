package gworker

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/pkg/errors"
)

type ErrorHandle func(err error, job WorkerJob)
type JobRunOverHandle func(worker Worker, job WorkerJob)

func HandleSignal() chan os.Signal {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT)
	signal.Notify(c, syscall.SIGTERM)
	return c
}

type Worker interface {
	IsBusy() bool
	Push(job WorkerJob) error
	Run(context.Context)
	GetTotalJob() int
	GetWorkerId() int
	Status() uint
	SetErrorHandle(ErrorHandle)
	SetJobRunOverHandle(JobRunOverHandle)
	PreSecondDealNum(num int)
}

type Gworker struct {
	cancelFunc        context.CancelFunc
	waitGroup         *sync.WaitGroup
	ctx               context.Context
	timeout           time.Duration
	waitTime          time.Duration
	errorHandel       ErrorHandle
	pool              WorkerPool
	job               chan WorkerJob
	stopFlag          bool
	status            uint
	workerId          int
	size              int
	runCastNs         int64
	preSecondDealNum  int
	currentSecondDeal int
	jobRunOverHandle  JobRunOverHandle
	attempt           int
}

func (w *Gworker) Attempt() int {
	return w.attempt
}

func (w *Gworker) OK() error {
	return nil
}

func (w *Gworker) Skip() error {
	return nil
}

func (w *Gworker) Retry(ctx context.Context, delay int) error {
	return nil
}

func (w *Gworker) Body() []byte {
	return []byte{}
}

func NewWorker(id int,
	timeout time.Duration,
	waitGroup *sync.WaitGroup,
	pool WorkerPool,
) *Gworker {
	jobSize := 100
	worker := &Gworker{
		workerId:    id,
		pool:        pool,
		timeout:     timeout,
		size:        jobSize,
		waitGroup:   waitGroup,
		errorHandel: pool.GetErrorHandle(),
		job:         make(chan WorkerJob, jobSize),
	}
	return worker
}

func (w *Gworker) PreSecondDealNum(num int) {
	w.preSecondDealNum = num
}

func (w *Gworker) Push(job WorkerJob) error {
	if w.stopFlag {
		return errors.New("worker stop")
	}
	if job == nil {
		return errors.New("job is nil")
	}
	w.job <- job
	return nil
}

func (w *Gworker) GetTotalJob() int {
	return len(w.job)
}

func (w *Gworker) IsBusy() bool {
	return len(w.job) == w.size
}

func (w *Gworker) GetWorkerId() int {
	return w.workerId
}

func (w *Gworker) Status() uint {
	return w.status
}

func (w *Gworker) GetErrorHandle() ErrorHandle {
	return w.errorHandel
}

func (w *Gworker) SetErrorHandle(handel ErrorHandle) {
	w.errorHandel = handel
}

func (w *Gworker) Run(ctx context.Context) {
	defer func() {
		// glog.Debugf("%d stop \n", w.workerId)
		if w.waitGroup != nil {
			w.waitGroup.Done()
		}
	}()

	for !w.stopFlag {
		select {
		case <-ctx.Done():
			w.stopFlag = true
			break
		case job := <-w.job:
			if job == nil {
				continue
			}
			func() {
				var err error
				defer func() {
					if w.jobRunOverHandle != nil {
						w.jobRunOverHandle(w, job)
					}

					if errInfo := recover(); errInfo != nil {
						if w.errorHandel != nil {
							info, _ := errInfo.(string)
							w.errorHandel(errors.New("recover from panic "+info), job)
						}
					}
				}()
				err = job.Run(ctx)
				if err != nil {
					w.errorHandel(err, job)
				}
			}()
			w.pool.RecycleWorker(w)
		}
	}
}

func (w *Gworker) SetJobRunOverHandle(jobRunOverHandle JobRunOverHandle) {
	if jobRunOverHandle == nil {
		return
	}
	w.jobRunOverHandle = jobRunOverHandle
}
