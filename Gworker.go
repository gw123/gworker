package gworker

import (
	"context"
	"github.com/pkg/errors"
	"sync"
	"time"
)

type Gworker struct {
	cancelFunc        context.CancelFunc
	waitGroup         *sync.WaitGroup
	ctx               context.Context
	timeout           time.Duration
	waitTime          time.Duration
	errorHandel       ErrorHandle
	pool              WorkerPool
	job               chan Job
	stopFlag          bool
	status            uint
	workerId          int
	size              int
	runOverTotal      int
	runCastNs         int64
	preSecondDealNum  int
	currentSecondDeal int
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
		job:         make(chan Job, jobSize),
	}
	return worker
}

func (w *Gworker) PreSecondDealNum(num int) {
	w.preSecondDealNum = num
}

func (w *Gworker) Push(job Job) error {
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
				defer func() {
					//防止job pnaic 影响协程继续正常工作
					if err := recover(); err != nil {
						w.errorHandel(errors.New("recover from panic"), job)
					}
				}()
				startTime := time.Now()
				err := job.Run()
				if err != nil {
					w.errorHandel(err, job)
				}
				w.runOverTotal++
				//fmt.Println(w.preSecondDealNum, w.currentSecondDeal)
				if w.preSecondDealNum > 0 {
					if w.currentSecondDeal < w.preSecondDealNum {
						endTime := time.Now()
						if startTime.Second() == endTime.Second() {
							w.currentSecondDeal++
						} else {
							w.currentSecondDeal = 0
						}
						w.runCastNs += endTime.Sub(startTime).Nanoseconds()
					} else {
						for ; ; {
							endTime := time.Now()
							if startTime.Second() == endTime.Second() {
								time.Sleep(time.Millisecond)
							}else{
								w.currentSecondDeal = 0
								break
							}
						}
					}
				}
			}()
			w.pool.RecycleWorker(w)
		}
	}
}
