package gworker

import (
	"context"
	"sync"
	"time"
)

type WorkerPool interface {
	Push(job Jobber) error
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
	Workers              []Worker
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

func NewWorkerPool(ctx context.Context, timeout time.Duration, poolSize int, errHandle ErrorHandle, overHandle JobRunOverHandle) *GworkerPool {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancelFunc := context.WithCancel(ctx)
	pool := &GworkerPool{
		ctx:              ctx,
		timeout:          timeout,
		Workers:          make([]Worker, 0),
		cancelFunc:       cancelFunc,
		poolSize:         poolSize,
		waitGroup:        &sync.WaitGroup{},
		freePool:         make(chan Worker, poolSize),
		errorHandel:      errHandle,
		jobRunOverHandle: overHandle,
	}
	return pool
}

func (pool *GworkerPool) init() {
	for i := 0; i < pool.poolSize; i++ {
		pool.waitGroup.Add(1)
		worker := NewWorker(i, pool.ctx, pool.timeout, pool.waitGroup, pool)
		worker.SetJobRunOverHandle(pool.jobRunOverHandle)
		worker.SetErrorHandle(pool.errorHandel)
		pool.freePool <- worker
		pool.Workers = append(pool.Workers, worker)
	}
	go pool.StartTimer()
}

func (pool *GworkerPool) Push(job Jobber) error {
	if pool.currentSecondDeal >= pool.preSecondDealNum {
		for ; pool.currentSecondDeal >= pool.preSecondDealNum; {
			time.Sleep(100 * time.Millisecond)
		}
	}
	worker := <-pool.freePool
	pool.currentSecondDeal++
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
	pool.init()
	for i := 0; i < pool.poolSize; i++ {
		go pool.Workers[i].Run()
	}
}

func (pool *GworkerPool) Stop() {
	if pool.stopFlag == true {
		return
	}
	pool.cancelFunc()
	//glog.Debug("wait all worker stop")
	pool.stopFlag = true
	for _, worker := range pool.Workers {
		worker.Stop()
	}
	pool.waitGroup.Wait()
	//glog.Debug("all run over")
	return
}

func (pool *GworkerPool) IsStop() bool {
	return pool.stopFlag
}

//func (pool *GworkerPool) Stop() chan int {
//	pool.cancelFunc()
//	glog.Debug("wait all worker stop")
//	for _, worker := range pool.Workers {
//		worker.Stop()
//	}
//	overCh := make(chan int, 1)
//	go func() {
//		fmt.Println("xxxxxxxxxxxxxxxxxxxxx")
//		pool.waitGroup.Wait()
//		fmt.Println("####################")
//		overCh <- 1
//	}()
//	return overCh
//}

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

func (pool *GworkerPool) SetJobRunOverHandle(handle JobRunOverHandle) {
	pool.jobRunOverHandle = handle
}

func (pool *GworkerPool) GetJobRunOverHandle() JobRunOverHandle {
	return pool.jobRunOverHandle
}

/***
设置每秒最大处理任务数量
*/
func (pool *GworkerPool) PreSecondDealNum(num int) {
	pool.preSecondDealNum = num
}
