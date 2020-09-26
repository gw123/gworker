package gworker

import (
	"context"
	"encoding/json"
	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/config"
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/pkg/errors"
	"time"
)

type Producer interface {
	PostJob(ctx context.Context, task Job) error
}

type ProducerManager struct {
	mqServer *machinery.Server
}

func NewPorducerManager(opt *Options) (*ProducerManager, error) {
	cfg := &config.Config{
		Broker:        opt.Broker,
		DefaultQueue:  opt.DefaultQueue,
		ResultBackend: opt.ResultBackend,
		AMQP: &config.AMQPConfig{
			Exchange:      opt.AMQP.Exchange,
			ExchangeType:  opt.AMQP.ExchangeType,
			PrefetchCount: opt.AMQP.PrefetchCount,
			AutoDelete:    opt.AMQP.AutoDelete,
		},
	}

	server, err := machinery.NewServer(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "create server")
	}

	return &ProducerManager{
		mqServer: server,
	}, nil
}

func (w *ProducerManager) PostJob(ctx context.Context, job Job) error {
	data, err :=  json.Marshal(job)
	if err != nil {
		return errors.Wrap(err, "job.marshalJson")
	}

	args := make([]tasks.Arg, 1)
	args[0] = tasks.Arg{
		Name:  "body",
		Type:  "string",
		Value: string(data),
	}

	signTask, err := tasks.NewSignature(job.GetName(), args)
	if err != nil {
		return errors.Wrap(err, "signature")
	}

	signTask.RoutingKey = job.GetName()
	signTask.RetryCount = job.RetryCount()

	if job.Delay() !=0 {
		eta := time.Now().UTC().Add(job.Delay())
		signTask.ETA =  &eta
	}


	_, err = w.mqServer.SendTaskWithContext(ctx, signTask)
	if err != nil {
		return errors.Wrap(err, "SendTaskWithContext")
	}
	return nil
}
