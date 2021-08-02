package rabbiter

import (
	"context"
	"github.com/gw123/gworker/amqpool"
	"sync"

	"github.com/streadway/amqp"
)

// HandlerFunc is function for handling job. The first paramater context.Context is a placeholder for future use.
type HandlerFunc func(ctx context.Context, job Jobber) error

// DefaultHandlerFunc is the default job handler.
func DefaultHandlerFunc(ctx context.Context, job Jobber) error {
	job.Logger().Warn("Using default handler function. Message is skipped.")
	return nil
}

// Consumer is used to consume message from a specific queue.
type Consumer interface {
	GetID() string
	IsRunning() bool
	Listen(ctx context.Context) error
	Stop()
	Wait() error
}

// ConsumeRabbiter implements Consumer interface.
type ConsumeRabbiter struct {
	*Rabbiter
	*amqpool.Consumer

	q *Queue
	e *Exchange
	h HandlerFunc
}

// openChannel is opener function for amqpool.Consumer.
func (r *ConsumeRabbiter) openChannel(ctx context.Context, pattern string) (amqpool.AMQPChannel, string, error) {
	ch, err := r.channel(ctx)
	if err != nil {
		return nil, "", err
	}

	if err := ch.Qos(r.q.PrefetchCount, 0, false); err != nil {
		return nil, "", err
	}

	if err := r.declareExchange(ch, r.e); err != nil {
		return nil, "", err
	}

	name, err := r.declareQueue(ch, r.q)
	if err != nil {
		return nil, "", err
	}

	if err := r.bindQueue(ch, name, pattern, r.e.Name); err != nil {
		return nil, "", err
	}

	return ch, name, nil
}

// handlerFunc receives a delivery, converts to a Job then calls user handler.
// This function will be asigned to underlying amqpool.Consumer's HandlerFunc attribute.
func (r *ConsumeRabbiter) handlerFunc(d *amqp.Delivery) error {
	job := r.newJob(d)

	err := r.h(context.Background(), job)
	if err != nil {
		r.logger.Errorf("consumer %s handler reported an error: %s", r.Consumer.ID, err)
	}

	// in case user didn't ack
	job.OK()

	// we handle messages' ack by ourself
	return amqpool.ErrHandlerHandle
}

// newJob receives a delivery, converts it to a Job.
func (r *ConsumeRabbiter) newJob(d *amqp.Delivery) *Job {
	attempt, ok := d.Headers[HeaderAttempt]
	if !ok {
		attempt = int32(0)
	}

	m := &Message{
		ID:          d.MessageId,
		Body:        d.Body,
		ContentType: d.ContentType,
		RoutingKey:  d.RoutingKey,
		Headers:     d.Headers,
		Attempt:     int(attempt.(int32)),
	}

	return &Job{
		Delivery: d,

		rabbiter: r.Rabbiter,
		message:  m,
		queue:    r.q,

		lock: new(sync.RWMutex),
	}
}
