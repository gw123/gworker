package gworker

import (
	"context"
	"github.com/gw123/glog"
	"sync"
	"time"
)

type GworkerPool struct {
	status               uint
	ctx                  context.Context
	timeout              time.Duration
	Workers              []Worker
	index                int
	cancelFunc           context.CancelFunc
	poolSize             int
	waitGroup            *sync.WaitGroup
	freePool             chan Worker
	workerStatusMapMutex sync.RWMutex
	errorHandel          ErrorHandle
	preSecondDealNum     int
}

func NewWorkerPool(ctx context.Context, timeout time.Duration, poolSize int, handle ErrorHandle) *GworkerPool {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancelFunc := context.WithCancel(ctx)
	pool := &GworkerPool{
		ctx:         ctx,
		timeout:     timeout,
		Workers:     make([]Worker, 0),
		cancelFunc:  cancelFunc,
		poolSize:    poolSize,
		waitGroup:   &sync.WaitGroup{},
		freePool:    make(chan Worker, poolSize),
		errorHandel: handle,
	}
	return pool
}

func (pool *GworkerPool) init() {
	for i := 0; i < pool.poolSize; i++ {
		pool.waitGroup.Add(1)
		worker := NewWorker(i, pool.ctx, pool.timeout, pool.waitGroup, pool)
		if pool.preSecondDealNum > 0 {
			worker.PreSecondDealNum(pool.preSecondDealNum)
		}
		pool.freePool <- worker
		pool.Workers = append(pool.Workers, worker)
	}
}

func (pool *GworkerPool) Push(job Job) error {
	worker := <-pool.freePool
	return worker.Push(job)
}

func (pool *GworkerPool) Run() {
	pool.init()
	for i := 0; i < pool.poolSize; i++ {
		go pool.Workers[i].Run()
	}
}

func (pool *GworkerPool) Stop() chan int {
	pool.cancelFunc()
	glog.Debug("wait all worker stop")
	for _, worker := range pool.Workers {
		worker.Stop()
	}
	overCh := make(chan int, 0)
	go func() {
		pool.waitGroup.Wait()
		overCh <- 1
	}()
	return overCh
}

/***
  在worker完成一个job后,将其放回free队列等待接收下一个任务
*/
func (pool *GworkerPool) RecycleWorker(worker Worker) {
	pool.freePool <- worker
}

func (pool *GworkerPool) Status() uint {
	return pool.status
}

func (pool *GworkerPool) GetErrorHandle() ErrorHandle {
	return pool.errorHandel
}

func (pool *GworkerPool) SetErrorHandle(handel ErrorHandle) {
	pool.errorHandel = handel
}

func (pool *GworkerPool) PreSecondDealNum(num int) {
	pool.preSecondDealNum = num
}
