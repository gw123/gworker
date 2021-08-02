package rabbiter

import (
	"context"
	"github.com/gw123/gworker/amqpool"
	"sync"

	"github.com/streadway/amqp"
)

const (
	// DefaultMaxConnections is default connection pool size.
	DefaultMaxConnections = 1

	// DefaultMaxChannels is default channel pool size.
	DefaultMaxChannels = 1

	// DefaultPrefetchCount is default prefetch count for consumer channel.
	DefaultPrefetchCount = 1
)

// Rabbiter implements Fpay's RabbitMQ usage. It implements publisher and consumer on top of RabbitMQ.
type Rabbiter struct {
	logger      amqpool.Logger
	connections *amqpool.ConnectionPool

	exchanges sync.Map
	queues    sync.Map
}

// Options is values to init a rabbiter.
type Options struct {
	Logger         amqpool.Logger
	Host           string
	MaxConnections int
	MaxChannels    int
	PrefetchCount  int
}

// NewRabbiter creates a new Rabbiter instance according to the options.
func NewRabbiter(options Options) *Rabbiter {
	logger := options.Logger
	if logger == nil {
		logger = amqpool.DefaultLogger
	}

	maxConn := options.MaxConnections
	if maxConn <= 0 {
		maxConn = DefaultMaxConnections
	}

	maxCh := options.MaxChannels
	if maxCh <= 0 {
		maxCh = DefaultMaxChannels
	}

	pool := amqpool.NewConnectionPool(
		maxConn,
		func(ctx context.Context) (*amqpool.Connection, error) {
			c := amqpool.NewConnection(amqpool.ConnectionOptions{
				Logger:      logger,
				MaxChannels: maxCh,
				Dial: func() (amqpool.AMQPConnection, error) {
					return amqp.Dial(options.Host)
				},
			})

			err := c.Connect(ctx)
			return c, err
		},
	)

	return &Rabbiter{
		logger:      logger,
		connections: pool,
	}
}

// Publisher returns a new publisher.
func (r *Rabbiter) Publisher() Publisher {
	return &PublishRabbiter{r}
}

// Consumer returns a new consumer.
func (r *Rabbiter) Consumer(e *Exchange, q *Queue, h HandlerFunc) Consumer {
	if q.PrefetchCount <= 0 {
		q.PrefetchCount = DefaultPrefetchCount
	}

	if h == nil {
		h = DefaultHandlerFunc
	}

	ac := &amqpool.Consumer{
		ID:      "",
		Logger:  r.logger,
		Pattern: q.RoutingKey,
	}

	c := ConsumeRabbiter{
		Rabbiter: r,
		Consumer: ac,
		q:        q,
		e:        e,
		h:        h,
	}

	ac.Open = c.openChannel
	ac.HandlerFunc = c.handlerFunc

	return c
}

func (r *Rabbiter) pooledChannel(ctx context.Context) (*amqpool.PooledChannel, error) {
	c, err := r.connections.Get(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Release()

	ch, err := c.Channel()
	if err != nil {
		return nil, err
	}

	return ch, nil
}

func (r *Rabbiter) channel(ctx context.Context) (*amqpool.Channel, error) {
	c, err := r.connections.Get(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Release()

	ch, err := c.Channel()
	if err != nil {
		return nil, err
	}
	ch.Free()

	return ch.Channel, nil
}

func (r *Rabbiter) declareExchange(ch amqpool.AMQPChannel, ex *Exchange) error {
	key := ex.Name

	if _, ok := r.exchanges.Load(key); ok {
		return nil
	}

	if err := ch.ExchangeDeclare(ex.Name, ex.Type, ex.Durable, ex.AutoDeleted, ex.Internal, ex.NoWait, ex.Args); err != nil {
		return err
	}

	r.exchanges.Store(key, ex)
	return nil
}

func (r *Rabbiter) declareQueue(ch amqpool.AMQPChannel, q *Queue) (string, error) {
	key := q.Name

	if v, ok := r.queues.Load(key); ok {
		return v.(*Queue).Name, nil
	}

	aq, err := ch.QueueDeclare(q.Name, q.Durable, q.AutoDeleted, q.Exclusive, q.NoWait, q.Args)
	if err != nil {
		return "", err
	}

	q.Name = aq.Name
	key = q.Name
	r.queues.Store(key, q)

	return q.Name, nil
}

func (r *Rabbiter) bindQueue(ch amqpool.AMQPChannel, queue, key, exchange string) error {
	return ch.QueueBind(queue, key, exchange, false, nil)
}

func (r *Rabbiter) declareDeferredQueue(ch amqpool.AMQPChannel, queue, exchange, key string, delay int) (string, error) {
	args := amqp.Table{
		"x-dead-letter-exchange":    exchange,
		"x-dead-letter-routing-key": key,
		"x-message-ttl":             int32(delay * 1000),
	}
	return r.declareQueue(ch, &Queue{Name: queue, Durable: true, Args: args})
}

func (r *Rabbiter) declareDeadLetterExchange(ch amqpool.AMQPChannel, exchange string) error {
	return r.declareExchange(ch, &Exchange{Name: exchange, Type: "direct", Durable: true})
}
