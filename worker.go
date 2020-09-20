package gworker

import (
	"context"
	"github.com/pkg/errors"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func HandleSignal() chan os.Signal {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT)
	signal.Notify(c, syscall.SIGTERM)
	return c
}

type Worker interface {
	IsBusy() bool
	Push(job Jobber) error
	Run()
	Stop()
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
	job               chan Jobber
	stopFlag          bool
	status            uint
	workerId          int
	size              int
	runOverTotal      int
	runCastNs         int64
	preSecondDealNum  int
	currentSecondDeal int
	jobRunOverHandle  JobRunOverHandle
}

func NewWorker(id int,
	parent context.Context,
	timeout time.Duration,
	waitGroup *sync.WaitGroup,
	pool WorkerPool,
) *Gworker {
	ctx, cancelFunc := context.WithCancel(parent)
	jobSize := 10
	worker := &Gworker{
		workerId:    id,
		ctx:         ctx,
		pool:        pool,
		stopFlag:    false,
		timeout:     timeout,
		size:        jobSize,
		waitGroup:   waitGroup,
		cancelFunc:  cancelFunc,
		errorHandel: pool.GetErrorHandle(),
		job:         make(chan Jobber, jobSize),
	}
	return worker
}

func (w *Gworker) PreSecondDealNum(num int) {
	w.preSecondDealNum = num
}

func (w *Gworker) Push(job Jobber) error {
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

func (w *Gworker) Stop() {
	w.stopFlag = true
	close(w.job)
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

func (w *Gworker) Run() {
	defer func() {
		//fmt.Printf("%d stop \n", w.workerId)
		if w.waitGroup != nil {
			w.waitGroup.Done()
		}
	}()

	for ; ; {
		select {
		case job, ok := <-w.job:
			if !ok {
				return
			}
			if job == nil {
				continue
			}
			func() {
				var err error
				defer func() {
					if w.jobRunOverHandle != nil {
						w.jobRunOverHandle(w, job)
					}
					//防止job pnaic 影响协程继续正常工作
					if errInfo := recover(); errInfo != nil {
						info, _ := errInfo.(string)
						w.errorHandel(errors.New("recover from panic "+info), job)
					}
				}()
				startTime := time.Now()
				err = job.Handle()
				if err != nil {
					w.errorHandel(err, job)
				}
				w.runOverTotal++
				endTime := time.Now()
				w.runCastNs += endTime.Sub(startTime).Nanoseconds()
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
