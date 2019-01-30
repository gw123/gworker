package main

import (
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"fmt"
	"github.com/gw123/gworker"
	"github.com/gw123/net_tool/net_utils"
	"github.com/gw123/gworker/jobs/netScanJob"
)

/****
 * 利用channel同步协程
 */
func main() {
	ipList := net_utils.GetIpList()
	group := gworker.NewWorkerGroup(50)
	group.Start()
	var falg_stop bool = false
	go func() {
		c := make(chan os.Signal)
		//signal.Notify(c, syscall.SIGUSR2)
		signal.Notify(c, syscall.SIGINT)
		signal.Notify(c, syscall.SIGSTOP)
		for {
			s := <-c
			//收到信号后的处理，这里只是输出信号内容，可以做一些更有意思的事
			fmt.Println("get signal:", s)
			fmt.Println("完成已处理任务队列后程序结束")
			falg_stop = true
			group.Stop()
		}
	}()

	group.Start()

	for _, ip := range ipList {
		job := netScanJob.NewScanJob(ip)
		group.DispatchJob(job)
	}

	group.WaitEmpty()
}
