package amqpool

import (
	"context"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

// Connection is struct holds amqp connection.
type Connection struct {
	Amqp     AMQPConnection
	Channels *ChannelPool

	dial   func() (AMQPConnection, error)
	logger Logger
	sleep  time.Duration

	mux       sync.RWMutex
	connected bool
	close     chan bool
}

// ConnectionOptions attributes set for init a Connection.
type ConnectionOptions struct {
	Dial          func() (AMQPConnection, error)
	Logger        Logger
	MaxChannels   int
	RetryInterval time.Duration
}

// NewConnection returns a new Connection pointer.
func NewConnection(options ConnectionOptions) *Connection {
	c := &Connection{
		dial:   options.Dial,
		logger: options.Logger,
		sleep:  options.RetryInterval,
		close:  make(chan bool),
	}

	if options.MaxChannels <= 0 {
		options.MaxChannels = DefaultChannels
	}

	pool := &ChannelPool{
		max:  options.MaxChannels,
		open: c.ChannelWithPool,
	}

	c.Channels = pool

	return c
}

// Connect function block dialing to RabbitMQ server until connection established.
func (c *Connection) Connect(ctx context.Context) error {
	var err error

	c.mux.Lock()
	if c.connected {
		c.mux.Unlock()
		return nil
	}

	c.logger.Printf("rabbitmq connection is establishing.")

	for {
		select {
		case <-c.close:
			c.mux.Unlock()
			return ErrConnectionStopped
		case <-ctx.Done():
			c.mux.Unlock()
			return ctx.Err()
		default:
		}

		c.Amqp, err = c.dial()
		if err != nil {
			time.Sleep(c.sleep)
		} else {
			break
		}
	}

	c.connected = true

	close := make(chan *amqp.Error)
	c.Amqp.NotifyClose(close)

	c.logger.Printf("rabbitmq connection is established.")

	// start another goroutine to watch the connection
	go c.watchAndReconnect(close)

	c.mux.Unlock()

	return nil
}

func (c *Connection) watchAndReconnect(notify <-chan *amqp.Error) {
	close := false

	select {
	case err, ok := <-notify:
		if !ok {
			close = true
		}
		c.logger.Printf("rabbitmq connection is closed: %s.", err)
	case <-c.close:
		close = true
	}

	c.mux.Lock()

	if c.Amqp != nil {
		c.Amqp.Close()
		c.Amqp = nil
	}
	c.connected = false
	c.Channels.clear()

	c.mux.Unlock()

	if close {
		return
	}

	c.Connect(context.Background())
}

// Close closes current connection.
func (c *Connection) Close() {
	c.close <- true
}

// IsConnected checks if the current connection is connected to RabbitMQ server.
func (c *Connection) IsConnected() bool {
	c.mux.RLock()
	defer c.mux.RUnlock()
	return c.connected
}

// Channel returns a PooledChannel from channel pool.
func (c *Connection) Channel() (*PooledChannel, error) {
	return c.Channels.Get()
}

// ChannelWithPool creates a new amqp.Channel wrapped by PooledChannel.
func (c *Connection) ChannelWithPool(pool *ChannelPool) (*PooledChannel, error) {
	c.mux.RLock()
	defer c.mux.RUnlock()

	ac, err := c.Amqp.Channel()
	if err != nil {
		return nil, err
	}

	channel := &PooledChannel{
		Channel: &Channel{
			AMQPChannel: ac,
			connected:   true,
		},
		pool: pool,
	}

	c.logger.Printf("new amqp channel created.")

	return channel, nil
}
