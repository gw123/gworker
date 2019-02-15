package gworker

import (
	"github.com/gw123/gworker/jobs"
	"sync"
	"fmt"
	"context"
)

/***
 * WorkerGroup
 */
type WorkerGroup struct {
	WorkerPipelines []*WorkerPipeline
	Length          uint32
	waitGroup       *sync.WaitGroup
	cancelFun       context.CancelFunc
	index           uint32
}

func NewWorkerGroup(size uint32) (*WorkerGroup) {
	this := new(WorkerGroup)
	this.WorkerPipelines = make([]*WorkerPipeline, size)
	this.Length = size
	this.waitGroup = &sync.WaitGroup{}
	this.index = 0
	rootCtx := context.Background()
	ctx, cancelFun := context.WithCancel(rootCtx)
	this.cancelFun = cancelFun

	for index, _ := range this.WorkerPipelines {
		this.WorkerPipelines[index] = NewWorker(this.waitGroup, true, ctx)
		this.WorkerPipelines[index].SetWorkerName(fmt.Sprintf("Worker_%d", index))
	}
	return this
}

func (this *WorkerGroup) DispatchJob(job jobs.Job) {
	for this.WorkerPipelines[this.index].IsBusy() {
		this.index += 1
		if this.index == this.Length {
			this.index = 0
		}
		//fmt.Printf("%d is busy\n", this.index)
	}
	//fmt.Printf("push %d \n", this.index)
	this.WorkerPipelines[this.index].AppendJob(job)
	this.index += 1
	if this.index == this.Length {
		this.index = 0
	}
}

func (this *WorkerGroup) Stop() {
	this.cancelFun()
}

func (this *WorkerGroup) Start() {
	for _, worker := range this.WorkerPipelines {
		worker.Begin()
	}
}

func (this *WorkerGroup) Wait() {
	this.waitGroup.Wait()
}

func (this *WorkerGroup) WaitEmpty() {
	for _, worker := range this.WorkerPipelines {
		worker.CancelLoopWait()
	}
	this.Wait()
}
