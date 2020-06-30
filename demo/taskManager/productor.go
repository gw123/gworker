package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/RichardKnop/machinery/v1/config"
	"github.com/gw123/glog"
	"github.com/gw123/gworker"
	"github.com/gw123/gworker/demo/taskManager/tasks"
)

func main() {
	//conf := &config.Config{
	//	Broker:        "amqp://gw:gao123456@localhost:5672/",
	//	DefaultQueue:  "machinery_tasks",
	//	ResultBackend: "redis://gao123456xyt!@127.0.0.1:6379",
	//	AMQP: &config.AMQPConfig{
	//		Exchange:      "machinery_exchange",
	//		ExchangeType:  "direct",
	//		BindingKey:    "machinery_task",
	//		PrefetchCount: 1,
	//	},
	//}
	//taskManager, err := gworker.NewTaskManager(conf, "task1")
	broker := ""
	queue := ""
	resultbackend := ""
	exchange := ""
	bindingKey := ""

	phone := ""
	code := ""
	total := 0
	flag.StringVar(&broker, "broker", "amqp: //root:123456@localhost:5672/", "broker")
	flag.StringVar(&queue, "queue", "test_task", "queue")
	flag.StringVar(&resultbackend, "result", "redis://123456@127.0.0.1:6379", "result backend")
	flag.StringVar(&exchange, "exchange", "test_task", "ampq exchange")
	flag.StringVar(&bindingKey, "binding", "test_task", "binding-key")
	flag.StringVar(&phone, "phone", "18611112222", "phone")
	flag.StringVar(&code, "code", "2010", "code")
	flag.IntVar(&total, "total", 1, "total")
	flag.Parse()

	conf := &config.Config{
		Broker:        broker,
		DefaultQueue:  queue,
		ResultBackend: resultbackend,
		AMQP: &config.AMQPConfig{
			Exchange:      exchange,
			ExchangeType:  "direct",
			BindingKey:    bindingKey,
			PrefetchCount: 1,
		},
	}

	taskManager, err := gworker.NewTaskManager(conf, "")
	if err != nil {
		glog.Errorf("NewTaskManager %s", err)
		return
	}

	ctx := context.Background()
	for i := 0; i < total; i++ {
		n := fmt.Sprintf("%s-%d", code, i)
		smstask := tasks.NewSMSTask(phone, []string{n})
		err = taskManager.SendTask(ctx, smstask)
		if err != nil {
			glog.Errorf("taskManager.SendTask(ctx, smstask) %s", err)
		}
	}
}
