## golang 实现多个worker去解决大量的任务

- 可以设置worker在队列空的时候结束
- 设置单worker每秒最大执行job数量
- 提供友好的协程同步控制机制
- 可以灵活的自定义job

## 使用方式

```
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
	var runOverTotal = 1
	var mutex sync.Mutex
	pool := gworker.NewWorkerPool(nil, time.Second*5, 100, func(err error, job gworker.Job) {
		fmt.Println("ErrorHandle " + err.Error())
	}, func(worker gworker.Worker, job gworker.Job) {
		mutex.Lock()
		runOverTotal ++
		fmt.Println("run over" , runOverTotal)
		mutex.Unlock()
	})
	pool.PreSecondDealNum(10)
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

```

## 函数方法

### workerPool

```
    //push job  into workerPool
    Push(job Job) error
    //开始运行wokerPool
   	Run()
   	//停止 ,这里可能会阻塞
   	Stop() 
   	//回收一个空闲的worker
   	RecycleWorker(worker Worker)
   	//获取状态
   	Status() uint
   	//获取出错执行的函数
   	GetErrorHandle() ErrorHandle
   	//设置出错回调函数(panic时候触发)
   	SetErrorHandle(ErrorHandle)
   	//设置每秒处理任务数量
   	PreSecondDealNum(num int)
    //workerPool 是否已经停止运行
    IsStop() bool
    //设置任务执行完成后的回调函数
	SetJobRunOverHandle(JobRunOverHandle)
    //获取任务执行完成回调
	GetJobRunOverHandle() JobRunOverHandle
```








