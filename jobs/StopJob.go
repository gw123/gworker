package jobs

import (
	"github.com/gw123/gworker/interfaces"
	"time"
)

type StopJob struct {
	Job
}

func NewStopJob() (job *StopJob) {
	job = new(StopJob)
	job.CreatedTime = time.Now().Unix()
	job.Flag = interfaces.JobFlagEnd
	job.runFlag = true
	return
}
