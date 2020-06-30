package mq

import (
	"context"
	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/config"
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/gw123/glog"
	"github.com/pkg/errors"
)

type Task interface {
	GetTaskName() string
	ToJson() string
	GetHandleFun() interface{}
}

type TaskManager struct {
	mqServer          *machinery.Server
	worker            *machinery.Worker
	errorHandler      func(err error)
	preTaskHandler    func(*tasks.Signature)
	postTaskHandler   func(*tasks.Signature)
	preConsumeHandler func(*TaskManager) bool
}

func NewTaskManager(cfg *config.Config, consumerTag string) (*TaskManager, error) {
	server, err := machinery.NewServer(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "create server")
	}
	worker := server.NewWorker(consumerTag, 0)
	worker.SetErrorHandler(func(err error) {
		glog.Errorf("Task Manager ErrorHandel %s", err.Error())
	})
	return &TaskManager{
		mqServer: server,
		worker:   worker,
	}, nil
}

func (w *TaskManager) RegisterTask(task Task) error {
	w.mqServer.RegisterTask(task.GetTaskName(), task.GetHandleFun())
	return nil
}

func (w *TaskManager) StartWork(comsumeTag string, num int) {
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

	w.SetErrorHandler(func(err error) {
		glog.Errorf("ErrorHander :%s", err)
	})
	w.SetPerTaskHandler(func(signature *tasks.Signature) {
		glog.Infof("Name : %s", signature.Name)
	})
	worker.Launch()
}

func (w *TaskManager) SetPostTaskHandler(postTaskHandler func(*tasks.Signature)) {
	w.postTaskHandler = postTaskHandler
}

func (w *TaskManager) SetErrorHandler(errorHandler func(err error)) {
	w.errorHandler = errorHandler
}

func (w *TaskManager) SetPerTaskHandler(preTaskHandler func(*tasks.Signature)) {
	w.preTaskHandler = preTaskHandler
}

func (w *TaskManager) SetPreConsumerHandler(preConsumeHandler func(*TaskManager) bool) {
	w.preConsumeHandler = preConsumeHandler
}

func (w *TaskManager) SendTask(ctx context.Context, task Task) error {
	args := make([]tasks.Arg, 1)
	args[0] = tasks.Arg{
		Name:  "body",
		Type:  "string",
		Value: task.ToJson(),
	}

	signTask, err := tasks.NewSignature(task.GetTaskName(), args)
	if err != nil {
		return errors.Wrap(err, "signature")
	}
	signTask.RetryCount = 5
	_, err = w.mqServer.SendTaskWithContext(ctx, signTask)
	if err != nil {
		return errors.Wrap(err, "SendTaskWithContext")
	}
	return nil
}
