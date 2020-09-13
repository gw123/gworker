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

func (j *MyJob) Handle() error {
	//fmt.Printf("GID:%d ,%s\n", GetGID(), j.data)
	return nil
}

func TestRunPoll(t *testing.T) {

	var runTotal = 0
	var total = 10 *0000
	mutex := sync.Mutex{}
	pool := NewWorkerPool(nil, time.Second*5, 1000, func(err error, job Tasker) {

	}, func(worker Worker, job Tasker) {
		mutex.Lock()
		runTotal++
		mutex.Unlock()
	})
	pool.PreSecondDealNum(10 *0000)
	pool.Run()

	for i := 1; i <= total; i++ {
		job := NewMyJob(fmt.Sprintf("id: %d", i))
		pool.Push(job)
	}

	pool.Stop()
	if runTotal != total {
		t.Errorf("push job num %d not equal run job num %d", total, runTotal)
	}
}
