package gworker

import (
	//"time"
	"github.com/gw123/gworker/jobs"
	"fmt"
	"context"
	"errors"
	"sync"
	"time"
)

const MaxJobs = 10

type WorkerPipeline struct {
	workerName string
	Jobs       chan jobs.Job

	//读到个channel写入数据 worker 开始执行job
	startChan chan int
	stopChan  chan int

	StopFlag      bool
	runFlag       bool
	firstRunFlag  bool
	isBusy        bool
	jobTimeout    uint32
	isBusyMutex   sync.Mutex
	timeoutChan   <-chan time.Time
	SyncWaitGroup *sync.WaitGroup
	isLoopWait    bool

	//任务的处理速率   采样10秒钟处情况
	dealRate float32
	ctx      context.Context
}

/***
 * 处理
 */
func NewWorker(waitGroup *sync.WaitGroup, isLoopWait bool, ctx context.Context) (workerPipeline *WorkerPipeline) {
	workerPipeline = new(WorkerPipeline)
	workerPipeline.Jobs = make(chan jobs.Job, MaxJobs)
	workerPipeline.startChan = make(chan int)
	workerPipeline.stopChan = make(chan int)

	workerPipeline.SyncWaitGroup = waitGroup
	workerPipeline.SyncWaitGroup.Add(1)
	workerPipeline.StopFlag = false
	workerPipeline.runFlag = false
	workerPipeline.firstRunFlag = false
	workerPipeline.isBusy = false
	workerPipeline.isLoopWait = isLoopWait
	workerPipeline.ctx = ctx
	return
}

func (this *WorkerPipeline) SetWorkerName(workerName string) {
	this.workerName = workerName
}

func (this *WorkerPipeline) GetWorkerName() string {
	return this.workerName
}

func (this *WorkerPipeline) SetLoopWait(flag bool) {
	this.isLoopWait = flag
}

func (this *WorkerPipeline) IsBusy() bool {
	this.isBusyMutex.Lock()
	defer this.isBusyMutex.Unlock()

	if len(this.Jobs) == MaxJobs {
		return true
	}
	return false
}

func (this *WorkerPipeline) SetTimeOut(timeout time.Duration) {
	this.timeoutChan = time.After(timeout)
}

func (this *WorkerPipeline) Stop() {
	this.StopFlag = true
	this.runFlag = false
	stopJob := jobs.NewStopJob()
	this.Jobs <- stopJob
}

func (this *WorkerPipeline) Start() {
	this.startChan <- 1
}

func (this *WorkerPipeline) Pause() {
	this.runFlag = false
	this.isBusy = true
}

func (this *WorkerPipeline) AppendJob(job jobs.Job) error {
	if this.StopFlag {
		fmt.Println(this.workerName, "AppendJob run over!!")
		return errors.New("流水线运行结束")
	}

	job.SetWorkerName(this.workerName)
	this.Jobs <- job
	return nil
}

func (this *WorkerPipeline) Begin() {
	if this.firstRunFlag {
		return
	}
	this.firstRunFlag = true
	this.runFlag = true
	go this.control()
	go this.run()
}

func (this *WorkerPipeline) control() {
	for ; ; {
		if this.StopFlag {
			break
		}
		select {
		case <-this.timeoutChan:
			fmt.Println("timeoutChan")
			this.Stop()
		case <-this.ctx.Done():
			fmt.Println("ctx.Done")
			this.Stop()
		}
	}
}

/***
  等待当前任务队列中所有结束
 */
func (this *WorkerPipeline) WaitEmpty() {
	this.CancelLoopWait()
	<-this.stopChan
}

/***
  等待当前任务队列中所有结束
 */
func (this *WorkerPipeline) CancelLoopWait() {
	this.isLoopWait = false
	stopJob := jobs.NewStopJob()
	this.Jobs <- stopJob
}

func (this *WorkerPipeline) run() {
	defer func() {
		this.StopFlag = true
		this.runFlag = false
		this.isBusy = false
		if this.SyncWaitGroup != nil {
			this.SyncWaitGroup.Done()
		}
		this.stopChan <- 0
		//fmt.Println(this.WorkerName + "任务执行完成step3")
	}()

	var job jobs.Job
	for ; ; {
		if len(this.Jobs) == 0 && !this.isLoopWait {
			// 一次性执行的任务 已经完成
			fmt.Println(this.workerName + "任务执行完成step1")
			break
		}

		if this.StopFlag {
			//结束任务
			fmt.Println(this.workerName+"任务执行完成step1", "StopFlag")
			break
		}

		if !this.runFlag {
			//任务暂停，等待唤醒
			fmt.Println(this.workerName+"任务暂停，等待唤醒", "StopFlag")
			<-this.startChan
		}

		//在使用中注意这里会阻塞
		job = <-this.Jobs
		if job.GetJobFlag() == jobs.JobFlagEnd {
			//fmt.Println(this.workerName+" 结束任务", "JobFlagEnd")
			break
		}

		job.DoJob()

		//开始执行任务...
		//fmt.Println("Job over wId: "+job.GetWorkerName(), " "+string(job.GetPayload()))
	}

	//fmt.Println(this.WorkerName + "任务执行完成step2")
}
