package netScanJob

import (
	"fmt"
	jobs2 "github.com/gw123/gworker/demo/jobs"
	"net"
	"strconv"
	"time"
)

type ScanJob struct {
	jobs2.BaseJob
}

func (this *ScanJob) GetJobType() string {
	panic("implement me")
}

func (this *ScanJob) Run() error {
	panic("implement me")
}

func (this *ScanJob) Stop() {
	panic("implement me")
}

func NewScanJob(ip string) (job *ScanJob) {
	job = new(ScanJob)
	job.CreatedTime = time.Now().Unix()
	job.Flag = jobs2.JobFlagNormal
	job.Payload = []byte(ip)
	return
}

func (this *ScanJob) DoJob() {
	//fmt.Println("执行任务：", this.WorkerName, string(this.Payload))
	ip := string(this.Payload)
	addr := ip + ":" + strconv.Itoa(9100)
	conn, err := net.DialTimeout("tcp", addr, time.Second*2)
	if err != nil {
		//fmt.Println(err)
		return
	}
	defer conn.Close()
	fmt.Println("DoJob ", addr, "连接成功")

}
