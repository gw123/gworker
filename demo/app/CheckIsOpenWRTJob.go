package app

import (
	"io"
	"time"
	"github.com/gw123/net_tool/worker/interfaces"
	"fmt"
		"net"
)

type CheckIsOpenWRTJob struct {
	WorkerName  string
	CreatedTime int64
	UpdatedTime int64
	Flag        int64
	JobType     string
	Payload     []byte
	Response    []byte
	Input       io.WriteCloser
	Output      io.ReadCloser
}

func (this *CheckIsOpenWRTJob) SetWriteCloser(input io.WriteCloser) {
	this.Input = input
}

func (this *CheckIsOpenWRTJob) SetReadCloser(ouput io.ReadCloser) {
	this.Output = ouput
}

func (this *CheckIsOpenWRTJob) GetWorkerName() string {
	return this.WorkerName
}

func (this *CheckIsOpenWRTJob) SetWorkerName(workername string) {
	this.WorkerName = workername
}

func (this *CheckIsOpenWRTJob) SetPayload(payload []byte) {
	this.Payload = payload
}

func (this *CheckIsOpenWRTJob) GetPayload() []byte {
	return this.Payload
}

func (this *CheckIsOpenWRTJob) GetCreatedTime() int64 {
	return this.CreatedTime
}

func (this *CheckIsOpenWRTJob) SetJobFlag(flag int64) {
	this.Flag = flag
}

func (this *CheckIsOpenWRTJob) GetJobFlag() int64 {
	return this.Flag
}

func (this *CheckIsOpenWRTJob) DoJob() {
	//fmt.Println("执行Job" + string(this.Payload))
	remoteAddr := string(this.Payload)

	conn, _:=  net.DialTimeout("tcp", remoteAddr, time.Second*3)
	if conn == nil {
		//fmt.Println(remoteAddr + "不可以抵达" + err.Error())
	}else{
		fmt.Println(remoteAddr + "可以抵达")
	}
}

func NewCheckIsOpenWRTJob(addr string) (job *CheckIsOpenWRTJob) {
	job = new(CheckIsOpenWRTJob)
	job.CreatedTime = time.Now().Unix()
	job.Payload = []byte(addr)
	job.Flag = interfaces.JobFlagNormal
	return
}
