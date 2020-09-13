package gworker

import (
	"context"
	"encoding/json"
	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/config"
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/pkg/errors"
)

type Producer interface {
	PostTask(ctx context.Context, task Task) error
}

type ProducerManager struct {
	mqServer *machinery.Server
}

func NewPorducerManager(cfg *config.Config) (*ProducerManager, error) {
	server, err := machinery.NewServer(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "create server")
	}

	return &ProducerManager{
		mqServer: server,
	}, nil
}

func (w *ProducerManager) PostTask(ctx context.Context, task Task) error {
	data, err :=  json.Marshal(task)
	if err != nil {
		return errors.Wrap(err, "task.marshalJson")
	}

	args := make([]tasks.Arg, 1)
	args[0] = tasks.Arg{
		Name:  "body",
		Type:  "string",
		Value: string(data),
	}

	signTask, err := tasks.NewSignature(task.GetName(), args)
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
