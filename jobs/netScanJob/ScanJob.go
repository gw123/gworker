package netScanJob

import (
	"time"
	"github.com/gw123/gworker/interfaces"
	"github.com/gw123/gworker/jobs"
	"fmt"
	"strconv"
	"net"
)

type ScanJob struct {
	jobs.Job
}

func NewScanJob(ip string) (job *ScanJob) {
	job = new(ScanJob)
	job.CreatedTime = time.Now().Unix()
	job.Flag = interfaces.JobFlagNormal
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
