package amqpool

import (
	"context"
	"time"

	"github.com/streadway/amqp"
)

const (
	// DefaultPrefetchCount is amqp channel qos default prefetch_count
	DefaultPrefetchCount = 1

	// DefaultConnections is the default number of connections
	DefaultConnections = 1

	// DefaultChannels is the default number of channels per connection
	DefaultChannels = 1

	// DefaultRetryInterval is the default interval for reconnecting
	DefaultRetryInterval = 500 * time.Millisecond
)

// AMQPConnection is interface for amqp.Connection
type AMQPConnection interface {
	Channel() (*amqp.Channel, error)
	Close() error
	NotifyClose(receiver chan *amqp.Error) chan *amqp.Error
}

// AMQPChannel is interface for amqp.Channel
type AMQPChannel interface {
	Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)
	Close() error
	ExchangeBind(destination, key, source string, noWait bool, args amqp.Table) error
	ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error
	Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
	QueueBind(name, key, exchange string, noWait bool, args amqp.Table) error
	QueueUnbind(name, key, exchange string, args amqp.Table) error
	QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error)
	Qos(prefetchCount, prefetchSize int, global bool) error
}

// PooledAMQPChannel is amqp channel managed by channel pool
type PooledAMQPChannel interface {
	AMQPChannel
	Release()
}

// HandlerFunc is event handler function
type HandlerFunc func(d *amqp.Delivery) error

// Listener interface
type Listener interface {
	GetID() string
	Listen(ctx context.Context) error
	Stop()
	Wait() error
}

// Errors
var (
	// ErrHandlerSkip is used by handler, telling consumer to skip this message.
	ErrHandlerSkip = &Error{"handler_skip", "Handler skipped this message"}

	// ErrHandlerRetry is used by handler, telling consumer to retry this message.
	ErrHandlerRetry = &Error{"handler_retry", "Handler want to retry this message"}

	// ErrHandlerHandle is used to tell consumer not handle this message.
	ErrHandlerHandle = &Error{"handler_handle", "Handler has handled this message"}

	// ErrConnectionStopped means connection is stopped by user.
	ErrConnectionStopped = &Error{"connection_stopped", "Connection is stopped by user"}

	// ErrConsumerClosed means consumer is closed by remote.
	ErrConsumerClosed = &Error{"consumer_closed", "Consumer connection is closed by remote"}

	// ErrConsumerStopped means consumer is stopped by user.
	ErrConsumerStopped = &Error{"consumer_stopped", "Consumer is stopped by user"}
)

// Error is amqpool error
type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() string {
	return e.Message
}
