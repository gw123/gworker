package jobs

import (
	"io"
	"fmt"
	"time"
)

const JobFlagEnd = 2
const JobFlagNormal = 1

type Job interface {
	GetPayload() []byte
	GetCreatedTime() int64
	GetJobFlag() int64
	SetWorkerName(string2 string)
	GetWorkerName() string
	SetWriteCloser(input io.WriteCloser)
	SetReadCloser(ouput io.ReadCloser)
	//执行任务
	DoJob()
	// 取消当前正在执行的任务
	Cancel()
}


func NewJob(payload []byte) (job *BaseJob) {
	job = new(BaseJob)
	job.CreatedTime = time.Now().Unix()
	job.Payload = payload
	job.Flag = JobFlagNormal
	job.runFlag = true
	return
}

func (this *BaseJob) SetWriteCloser(input io.WriteCloser) {
	this.Input = input
}

func (this *BaseJob) SetReadCloser(ouput io.ReadCloser) {
	this.Output = ouput
}

func (this *BaseJob) GetWorkerName() string {
	return this.WorkerName
}

func (this *BaseJob) SetWorkerName(workername string) {
	this.WorkerName = workername
}

func (this *BaseJob) GetPayload() []byte {
	return this.Payload
}

func (this *BaseJob) GetCreatedTime() int64 {
	return this.CreatedTime
}

func (this *BaseJob) GetJobFlag() int64 {
	return this.Flag
}

func (this *BaseJob) DoJob() {
	fmt.Println("执行任务：", this.WorkerName, string(this.Payload))
}

func (this *BaseJob) Cancel() {
	this.runFlag = true
}
