package main

import (
	"flag"
	"github.com/gw123/glog"
	"github.com/gw123/gworker"
	smsTask "github.com/gw123/gworker/demo/taskManager"
)

func main() {
	broker := ""
	queue := ""
	resultbackend := ""
	exchange := ""
	bindingKey := ""
	consumerTag := ""
	workerNum := 0
	flag.StringVar(&broker, "broker", "", "broker")
	flag.StringVar(&queue, "queue", "sms", "queue")
	flag.StringVar(&resultbackend, "result", "redis://123456@127.0.0.1:6379", "result backend")
	flag.StringVar(&exchange, "exchange", "sms", "ampq exchange")
	flag.StringVar(&bindingKey, "binding", "sms", "binding-key")
	flag.Parse()

	conf := &gworker.Options{
		Broker:        broker,
		DefaultQueue:  queue,
		ResultBackend: resultbackend,
		AMQP: &gworker.AMQPOptions{
			Exchange:      exchange,
			ExchangeType:  "direct",
			BindingKey:    bindingKey,
			PrefetchCount: 1,
		},
	}
	taskManager, err := gworker.NewConsumer(conf, consumerTag)
	if err != nil {
		glog.Errorf("NewTaskManager : %s", err.Error())
		return
	}

	taskManager.RegisterTask(&smsTask.SMSTask{})
	taskManager.StartWork(consumerTag, workerNum)
}
