package gworker

import (
	"bytes"
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"
)

func GetGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

type MyJob struct {
	data string
}

func NewMyJob(data string) *MyJob {
	return &MyJob{
		data: data,
	}
}

func (j *MyJob) GetJobType() string {
	return "myjob"
}

func (j *MyJob) Run() error {
	fmt.Printf("GID:%d ,%s\n", GetGID(), j.data)
	return nil
}

func (j *MyJob) Stop() {
}

type ErrJob struct {
	MyJob
}

func NewErrJob(data string) *ErrJob {
	return &ErrJob{
		MyJob{data: data},
	}
}

func (j *ErrJob) Run() error {
	//fmt.Printf("GID:%d ,%s\n", GetGID(), j.data)
	panic("test painc")
	return nil
}

func CreatedJob() *ErrJob {
	return NewErrJob("Job Creted : " + time.Now().Format("15:04:05"))
}

func TestRunPoll(t *testing.T) {

	var runTotal = 0
	var total = 1000000
	mutex := sync.Mutex{}
	pool := NewWorkerPool(nil, time.Second*5, 1000, func(err error, job Job) {
		//fmt.Println("ErrorHandle " + err.Error())
		mutex.Lock()
		runTotal++
		mutex.Unlock()
	})
	pool.PreSecondDealNum(1000)
	pool.Run()

	for i := 1; i <= total; i++ {
		job := CreatedJob()
		pool.Push(job)
	}


	select {
	case <-pool.Stop():
		break
	}

	if runTotal != total {
		t.Errorf("push job num %d not equal run job num %d", total, runTotal)
	}
}
