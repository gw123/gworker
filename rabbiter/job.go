package rabbiter

import (
	"context"
	"github.com/gw123/gworker/amqpool"
	"sync"

	"github.com/streadway/amqp"
)

// Jobber
type Jobber interface {
	Attempt() int
	Message() *Message
	OK() error
	Skip() error
	Retry(ctx context.Context, delay int) error
	Logger() amqpool.Logger
}

// Job is an abstract layer for consuming message.
type Job struct {
	*amqp.Delivery

	rabbiter *Rabbiter
	queue    *Queue

	lock  *sync.RWMutex
	acked bool

	message *Message
}

// OK ack current message.
func (j *Job) OK() error {
	j.lock.RLock()
	if j.acked {
		j.lock.RUnlock()
		return nil
	}
	j.lock.RUnlock()

	if err := j.Ack(false); err != nil {
		return err
	}

	j.lock.Lock()
	j.acked = true
	j.lock.Unlock()

	return nil
}

// Skip nack current message.
func (j *Job) Skip() error {
	j.lock.RLock()
	if j.acked {
		j.lock.RUnlock()
		return nil
	}
	j.lock.RUnlock()

	if err := j.Nack(false, false); err != nil {
		return err
	}

	j.lock.Lock()
	j.acked = true
	j.lock.Unlock()

	return nil
}

// Retry sends message to a deferred queue to try this message after the delay seconds.
func (j *Job) Retry(ctx context.Context, delay int) error {
	if err := j.Skip(); err != nil {
		return err
	}
	return j.rabbiter.Publisher().PublishWithDelay(ctx, j.queue, j.message, delay)
}

// Attempt returns message's trying times.
func (j *Job) Attempt() int {
	return j.message.Attempt
}

// Message returns current job message.
func (j *Job) Message() *Message {
	return j.message
}

// Logger returns logger on rabbiter.
func (j *Job) Logger() amqpool.Logger {
	return j.rabbiter.logger
}
