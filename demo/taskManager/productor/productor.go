package main

import (
	"context"
	"encoding/json"
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

	phone := ""
	code := ""
	total := 0
	comId := 0
	flag.StringVar(&broker, "broker", "amqp://xyt:123@qys:5672/", "broker")
	flag.StringVar(&queue, "queue", "sms", "queue")
	flag.StringVar(&resultbackend, "result", "redis://123456@127.0.0.1:6379", "result backend")
	flag.StringVar(&exchange, "exchange", "sms", "ampq exchange")
	flag.StringVar(&bindingKey, "binding", "sms", "binding-key")
	flag.StringVar(&phone, "phone", "18611112222", "phone")
	flag.StringVar(&code, "code", "2010", "code")
	flag.IntVar(&total, "total", 1, "total")
	flag.IntVar(&comId, "com_id", 0, "com_id")
	flag.Parse()

	conf := &gworker.Options{
		Broker:        broker,
		DefaultQueue:  queue,
		ResultBackend: resultbackend,
		AMQP: &gworker.AMQPOptions{
			Exchange:      queue,
			ExchangeType:  "direct",
			BindingKey:    queue,
			PrefetchCount: 1,
		},
	}

	taskManager, err := gworker.NewPorducerManager(conf)
	if err != nil {
		glog.Errorf("NewTaskManager %s", err)
		return
	}

	ctx := context.Background()
	for i := 0; i < total; i++ {
		params := make(map[string]string)
		params["code"] = code
		data, _ := json.Marshal(params)
		smstask := smsTask.NewSMSTask(uint(comId), "aliyun", "login", []string{phone}, string(data))
		smstask.Type = "ticketCode"
		err = taskManager.PostJob(ctx, smstask)
		if err != nil {
			glog.Errorf("taskManager.SendTask(ctx, smstask) %s", err)
		}
	}
}
