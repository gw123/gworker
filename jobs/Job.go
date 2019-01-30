package jobs

import (
	"io"
	"time"
	"github.com/gw123/gworker/interfaces"
	"fmt"
)

type Job struct {
	WorkerName  string
	CreatedTime int64
	UpdatedTime int64
	Flag        int64
	JobType     string
	Payload     []byte
	Response    []byte
	Input       io.WriteCloser
	Output      io.ReadCloser
	runFlag     bool
}

func NewJob(payload []byte) (job *Job) {
	job = new(Job)
	job.CreatedTime = time.Now().Unix()
	job.Payload = payload
	job.Flag = interfaces.JobFlagNormal
	job.runFlag = true
	return
}

func (this *Job) SetWriteCloser(input io.WriteCloser) {
	this.Input = input
}

func (this *Job) SetReadCloser(ouput io.ReadCloser) {
	this.Output = ouput
}

func (this *Job) GetWorkerName() string {
	return this.WorkerName
}

func (this *Job) SetWorkerName(workername string) {
	this.WorkerName = workername
}

func (this *Job) GetPayload() []byte {
	return this.Payload
}

func (this *Job) GetCreatedTime() int64 {
	return this.CreatedTime
}

func (this *Job) GetJobFlag() int64 {
	return this.Flag
}

func (this *Job) DoJob() {
	fmt.Println("执行任务：", this.WorkerName, string(this.Payload))
}

func (this *Job) Cancel() {
	this.runFlag = true
}
