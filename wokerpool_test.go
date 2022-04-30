package gworker

import (
	"bytes"
	"context"
	"runtime"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gw123/glog"
)

var runTotal int64

func GetGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

type MyJob struct {
	data []byte
}

func (j *MyJob) Run(ctx context.Context) error {
	atomic.AddInt64(&runTotal, 1)
	return nil
}

func (j *MyJob) Stop() {
	return
}

func NewMyJob(data []byte) *MyJob {
	return &MyJob{
		data: data,
	}
}

func TestRunPoll(t *testing.T) {
	var total int64 = 10000000
	var i int64 = 0
	pool := NewWorkerPool(nil, time.Second*5, 6)

	pool.PreSecondDealNum(1000 * 0000)
	pool.Run()

	for ; i < total; i++ {
		job := NewMyJob(nil)
		pool.Push(job)
	}

	pool.Stop()
	glog.Infof(" %d --- %d", total, runTotal)
	if runTotal != total {
		t.Errorf("push job num %d not equal run job num %d", total, runTotal)
	}
}
