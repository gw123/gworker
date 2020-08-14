package main

import (
	"flag"
	"github.com/RichardKnop/machinery/v1/config"
	"github.com/gw123/glog"
	"github.com/gw123/gworker/task"
)

func main() {
	broker := ""
	queue := ""
	resultbackend := ""
	exchange := ""
	bindingKey := ""
	consumerTag := ""
	workerNum := 0
	flag.StringVar(&broker, "broker", "amqp: //gw:gao123456@localhost:5672/", "broker")
	flag.StringVar(&queue, "queue", "test_task", "queue")
	flag.StringVar(&resultbackend, "result", "redis://123456@127.0.0.1:6379", "result")
	flag.StringVar(&exchange, "exchange", "test_task", "ampq exchange")
	flag.StringVar(&bindingKey, "binding", "test_task", "binding-key")
	flag.StringVar(&consumerTag, "consumer", "consume_01", "consumeTag")
	flag.IntVar(&workerNum, "worker", 0, "worker num")
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
	taskManager, err := task.NewTaskManager(conf, consumerTag)
	if err != nil {
		glog.Errorf("NewTaskManager : %s", err.Error())
		return
	}

	taskManager.RegisterTask(&SMSTask{})
	taskManager.StartWork(consumerTag, workerNum)
}
