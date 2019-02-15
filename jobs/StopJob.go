package jobs

import (
	"time"
)

type StopJob struct {
	BaseJob
}

func NewStopJob() (job *StopJob) {
	job = new(StopJob)
	job.CreatedTime = time.Now().Unix()
	job.Flag = JobFlagEnd
	job.runFlag = true
	return
}
