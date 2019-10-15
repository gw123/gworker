package main

import (
	"bytes"
	"fmt"
	"github.com/gw123/gworker"
	"runtime"
	"strconv"
	"sync"
	"time"
)

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

func GetGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

func (j *MyJob) Run() ( error) {
	fmt.Printf("GID:%d ,%s\n", GetGID(), j.data)
	//panic("test panic")
	return nil
}

func (j *MyJob) Stop() {
}

func CreatedJob() *MyJob {
	return NewMyJob("Job Creted : " + time.Now().Format("15:04:05"))
}

func main() {
	var runOverTotal = 0
	var mutex sync.Mutex
	pool := gworker.NewWorkerPool(nil, time.Second*5, 100, func(err error, job gworker.Job) {
		fmt.Println("ErrorHandle " + err.Error())
	}, func(worker gworker.Worker, job gworker.Job) {
		mutex.Lock()
		runOverTotal ++
		fmt.Println("run over" , runOverTotal)
		mutex.Unlock()
	})
	pool.PreSecondDealNum(1000)
	pool.Run()

	go func() {
		select {
		case <-gworker.HandleSignal():
			break
		}
		pool.Stop()
	}()

	startTime := time.Now()
	for i := 0; i < 100000&& !pool.IsStop(); i++ {
		job := CreatedJob()
		pool.Push(job)
	}

	pool.Stop()
	endTime := time.Now()
	fmt.Printf("cast time %d \n", endTime.Sub(startTime).Nanoseconds()/1000000)
}
