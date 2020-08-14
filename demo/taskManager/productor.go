package main

import (
	"context"
	"encoding/json"
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

	phone := ""
	code := ""
	total := 0
	comId := 0
	flag.StringVar(&broker, "broker", "amqp: //root:123456@localhost:5672/", "broker")
	flag.StringVar(&queue, "queue", "test_task", "queue")
	flag.StringVar(&resultbackend, "result", "redis://123456@127.0.0.1:6379", "result backend")
	flag.StringVar(&exchange, "exchange", "test_task", "ampq exchange")
	flag.StringVar(&bindingKey, "binding", "test_task", "binding-key")
	flag.StringVar(&phone, "phone", "18611112222", "phone")
	flag.StringVar(&code, "code", "2010", "code")
	flag.IntVar(&total, "total", 1, "total")
	flag.IntVar(&comId, "com_id", 0, "com_id")
	flag.Parse()

	conf := &config.Config{
		Broker:        broker,
		DefaultQueue:  queue,
		ResultBackend: resultbackend,
		AMQP: &config.AMQPConfig{
			Exchange:      queue,
			ExchangeType:  "direct",
			BindingKey:    queue,
			PrefetchCount: 1,
		},
	}

	taskManager, err := task.NewTaskManager(conf, "")
	if err != nil {
		glog.Errorf("NewTaskManager %s", err)
		return
	}

	ctx := context.Background()
	for i := 0; i < total; i++ {
		params := make(map[string]string)
		params["code"] = code
		data, _ := json.Marshal(params)
		smstask := NewSMSTask(uint(comId), "aliyun", "login", []string{phone}, string(data))
		smstask.Type = "loginCode"
		err = taskManager.SendTask(ctx, smstask)
		if err != nil {
			glog.Errorf("taskManager.SendTask(ctx, smstask) %s", err)
		}
	}
}
