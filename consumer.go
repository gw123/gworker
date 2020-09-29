package gworker

import (
	"encoding/json"
	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/config"
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/gw123/glog"
	"github.com/pkg/errors"
)

type Consumer interface {
	StartWork(comsumeTag string, num int)
	SetPostTaskHandler(postTaskHandler func(*tasks.Signature))
	SetErrorHandler(errorHandler func(err error))
	SetPerTaskHandler(preTaskHandler func(*tasks.Signature))
	SetPreConsumerHandler(preConsumeHandler func(*Consumer) bool)
}

type ConsumerManager struct {
	mqServer          *machinery.Server
	worker            *machinery.Worker
	errorHandler      func(err error)
	preTaskHandler    func(*tasks.Signature)
	postTaskHandler   func(*tasks.Signature)
	preConsumeHandler func(*Consumer) bool
}

func NewConsumer(opt * Options, jobber Jobber) (*ConsumerManager, error) {
	cfg := &config.Config{
		Broker:        opt.Broker,
		DefaultQueue:  jobber.GetName(),
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
	worker := server.NewWorker(jobber.GetName(), 0)
	worker.SetErrorHandler(func(err error) {
		glog.Errorf("Task Manager ErrorHandel %s", err.Error())
	})

	m := &ConsumerManager{
		mqServer: server,
		worker:   worker,
	}
	m.registerTask(jobber)
	return m ,nil
}

func (w *ConsumerManager) registerTask(job Jobber) error {
	w.mqServer.RegisterTask(job.GetName(), func(data string) error {
		err := json.Unmarshal([]byte(data), job)
		if err != nil{
			return err
		}
		return job.Handle()
	})
	return nil
}

func (w *ConsumerManager) StartWork(comsumeTag string, num int) {
	worker := w.mqServer.NewWorker(comsumeTag, num)
	if w.errorHandler != nil {
		worker.SetErrorHandler(w.errorHandler)
	}

	if w.preTaskHandler != nil {
		worker.SetPreTaskHandler(w.preTaskHandler)
	}

	if w.preConsumeHandler != nil {
		w.SetPreConsumerHandler(w.preConsumeHandler)
	}

	if w.errorHandler != nil {
		w.SetErrorHandler(w.errorHandler)
	}

	w.SetPerTaskHandler(func(signature *tasks.Signature) {
		glog.Infof("Name : %s", signature.Name)
	})
	worker.Launch()
}

func (w *ConsumerManager) SetPostTaskHandler(postTaskHandler func(*tasks.Signature)) {
	w.postTaskHandler = postTaskHandler
}

func (w *ConsumerManager) SetErrorHandler(errorHandler func(err error)) {
	w.errorHandler = errorHandler
}

func (w *ConsumerManager) SetPerTaskHandler(preTaskHandler func(*tasks.Signature)) {
	w.preTaskHandler = preTaskHandler
}

func (w *ConsumerManager) SetPreConsumerHandler(preConsumeHandler func(*Consumer) bool) {
	w.preConsumeHandler = preConsumeHandler
}
