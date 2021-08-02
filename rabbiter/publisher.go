package rabbiter

import (
	"context"
	"strconv"

	"github.com/streadway/amqp"
)

// Publisher is used to publish message to RabbitMQ.
type Publisher interface {
	Publish(ctx context.Context, ex *Exchange, m *Message) error
	PublishWithDelay(ctx context.Context, q *Queue, m *Message, delay int) error
}

// PublishRabbiter implements Publisher interface.
type PublishRabbiter struct {
	*Rabbiter
}

// Publish sends message through exchange.
func (r *PublishRabbiter) Publish(ctx context.Context, ex *Exchange, m *Message) error {
	ch, err := r.pooledChannel(ctx)
	if err != nil {
		return err
	}
	defer ch.Release()

	if err := r.declareExchange(ch, ex); err != nil {
		return err
	}

	return ch.Publish(
		ex.Name,
		m.RoutingKey,
		false,
		false,
		amqp.Publishing{
			MessageId:    m.ID,
			DeliveryMode: amqp.Persistent,
			ContentType:  m.ContentType,
			Body:         m.Body,
			Headers: amqp.Table{
				HeaderAttempt: int32(0),
			},
		},
	)
}

// PublishWithDelay sends a message directly to a queue with a delay.
func (r *PublishRabbiter) PublishWithDelay(ctx context.Context, q *Queue, m *Message, delay int) error {
	ch, err := r.pooledChannel(ctx)
	if err != nil {
		return err
	}
	defer ch.Release()

	dq := r.deferredQueueName(q.Name, m, delay)
	dex := dq

	if delay > 0 {
		if _, err := r.declareDeferredQueue(ch, dq, dex, m.RoutingKey, delay); err != nil {
			return err
		}

		if err := r.declareDeadLetterExchange(ch, dex); err != nil {
			return err
		}

		if err := r.bindQueue(ch, q.Name, m.RoutingKey, dex); err != nil {
			return err
		}
	}

	headers := m.Headers
	if headers == nil {
		headers = make(amqp.Table)
	}
	headers[HeaderAttempt] = int32(m.Attempt + 1)

	// send message to the deferred queue
	return ch.Publish(
		"",
		dq,
		false,
		false,
		amqp.Publishing{
			MessageId:    m.ID,
			DeliveryMode: amqp.Persistent,
			ContentType:  m.ContentType,
			Body:         m.Body,
			Headers:      headers,
		},
	)
}

func (r *PublishRabbiter) deferredQueueName(name string, m *Message, delay int) string {
	if len(m.RoutingKey) > 0 {
		name += "." + m.RoutingKey
	}

	if delay > 0 {
		name += ".deferred_" + strconv.Itoa(delay)
	}

	return name
}
