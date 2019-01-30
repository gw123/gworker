package main

import (
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"fmt"
	"github.com/gw123/gworker"
	"github.com/gw123/gworker/jobs"
	"runtime"
)

/****
 * 利用channel同步协程
 */
func main() {
	runtime.GOMAXPROCS(10)

	var falg_stop bool = false
	go func() {
		c := make(chan os.Signal)
		//signal.Notify(c, syscall.SIGUSR2)
		signal.Notify(c, syscall.SIGINT)
		for {
			s := <-c
			//收到信号后的处理，这里只是输出信号内容，可以做一些更有意思的事
			fmt.Println("get signal:", s)
			fmt.Println("完成已处理任务队列后程序结束")
			falg_stop = true
		}
	}()

	group := gworker.NewWorkerGroup(10)
	group.Start()

	for i := 1; i <= 10000000; i++ {
		job := jobs.NewJob([]byte(fmt.Sprintf("job %d", i)))
		group.DispatchJob(job)
	}

	group.WaitEmpty()
}
