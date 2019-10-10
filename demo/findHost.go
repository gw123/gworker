package main

import (
	"fmt"
	"github.com/gw123/gworker"
	"github.com/gw123/gworker/demo/jobs/netScanJob"
	"github.com/gw123/net_tool/netInterfaces"
	_ "net/http/pprof"
)

/****
 * 利用channel同步协程
 */
func main() {
	ipList, _, err := netInterfaces.GetIpList(false)
	if err != nil {
		fmt.Println(err)
		return
	}

	group := old.NewWorkerGroup(50)
	group.Start()

	go func() {
		select {
		case <-gworker.HandleSignal():
			break
		}
		group.Stop()
	}()

	for _, ip := range ipList {
		job := netScanJob.NewScanJob(ip)
		group.DispatchJob(job)
	}

	group.WaitEmpty()
}
