package gworker

import (
	"context"
	"sync"
	"time"
)

type WorkerPool interface {
	Push(job WorkerJob) error
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

type GworkerPool struct {
	status               uint
	ctx                  context.Context
	timeout              time.Duration
	index                int
	cancelFunc           context.CancelFunc
	poolSize             int
	waitGroup            *sync.WaitGroup
	freePool             chan Worker
	workerStatusMapMutex sync.RWMutex
	errorHandel          ErrorHandle
	preSecondDealNum     int
	currentSecondDeal    int
	preDealLock          sync.Mutex
	stopFlag             bool
	jobRunOverHandle     JobRunOverHandle
}

func NewWorkerPool(ctx context.Context, timeout time.Duration, poolSize int) *GworkerPool {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancelFunc := context.WithCancel(ctx)
	pool := &GworkerPool{
		ctx:              ctx,
		timeout:          timeout,
		cancelFunc:       cancelFunc,
		poolSize:         poolSize,
		waitGroup:        &sync.WaitGroup{},
		freePool:         make(chan Worker, poolSize),
		preSecondDealNum: 100000,
	}
	return pool
}

func (pool *GworkerPool) Push(job WorkerJob) error {
	//if pool.currentSecondDeal >= pool.preSecondDealNum {
	//	for pool.currentSecondDeal >= pool.preSecondDealNum {
	//		time.Sleep(100 * time.Millisecond)
	//	}
	//}
	worker := <-pool.freePool
	// pool.currentSecondDeal++
	return worker.Push(job)
}

func (pool *GworkerPool) StartTimer() {
	ticker := time.NewTicker(time.Second)
	go func() {
		for {
			<-ticker.C
			pool.currentSecondDeal = 0
		}
	}()
}

func (pool *GworkerPool) Run() {
	for i := 0; i < pool.poolSize; i++ {
		pool.waitGroup.Add(1)
		worker := NewWorker(i, pool.timeout, pool.waitGroup, pool)
		go worker.Run(pool.ctx)
		pool.freePool <- worker
	}
	go pool.StartTimer()
}

func (pool *GworkerPool) Stop() {
	if pool.stopFlag == true {
		return
	}
	pool.cancelFunc()
	pool.stopFlag = true
	pool.waitGroup.Wait()
	return
}

func (pool *GworkerPool) IsStop() bool {
	return pool.stopFlag
}

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

func (pool *GworkerPool) SetJobRunOverHandle(handle JobRunOverHandle) {
	pool.jobRunOverHandle = handle
}

func (pool *GworkerPool) GetJobRunOverHandle() JobRunOverHandle {
	return pool.jobRunOverHandle
}

func (pool *GworkerPool) PreSecondDealNum(num int) {
	pool.preSecondDealNum = num
}
