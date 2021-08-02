package amqpool

import (
	"context"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

// Consumer is amqp queue consumer
type Consumer struct {
	ID          string
	Logger      Logger
	HandlerFunc HandlerFunc
	Pattern     string
	Open        func(ctx context.Context, pattern string) (AMQPChannel, string, error)

	stop chan bool

	mu       sync.RWMutex
	running  bool
	waitings []chan error
}

// Listen start the consumer. If failed to connect to the amqp broker on the first start,
// it returns the connection error.
func (c *Consumer) Listen(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.running {
		return nil
	}

	// acquire a channel to consume
	ch, q, err := c.Open(ctx, c.Pattern)
	if err != nil {
		return err
	}

	tag := c.ID
	input, err := ch.Consume(q, tag, false, false, false, false, nil)
	if err != nil {
		return err
	}

	c.running = true
	c.stop = make(chan bool)

	go c.do(ch, input)

	c.Logger.Printf("listener %s is running\n", c.ID)

	return nil
}

func (c *Consumer) do(ch AMQPChannel, input <-chan amqp.Delivery) {
	var err error
	for flag := true; flag; {
		select {
		case d, ok := <-input:
			if ok {
				c.handle(&d)
			} else {
				flag = false
				err = ErrConsumerClosed
			}
		case <-c.stop:
			flag = false
			err = ErrConsumerStopped
		}
	}

	c.mu.Lock()
	c.running = false
	close(c.stop)
	c.mu.Unlock()

	// close channel after consumer stops
	ch.Close()

	c.release(err)

	// handle auto reconnecting
	if err == ErrConsumerClosed {
		c.Logger.Printf("listener %s stopped unexpectly: %s\n", c.ID, err)
		c.reconnect()
	}
}

func (c *Consumer) reconnect() {
	c.Logger.Printf("listener %s attempt reconnecting\n", c.ID)

	if err := c.Listen(context.Background()); err != nil {
		time.AfterFunc(DefaultRetryInterval, c.reconnect)
		return
	}

	c.Logger.Printf("listener %s resumed\n", c.ID)
}

func (c *Consumer) handle(d *amqp.Delivery) {
	err := c.HandlerFunc(d)
	switch err {
	case nil:
		d.Ack(false)
	case ErrHandlerSkip:
		d.Nack(false, false)
	case ErrHandlerRetry:
		d.Nack(false, true)
	case ErrHandlerHandle:
		return
	default:
		d.Nack(false, true)
		c.Logger.Printf("listener %s handler reported an unknown error: %s", c.ID, err)
	}
}

// Stop telling a running listener stop consuming
func (c *Consumer) Stop() {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.running {
		c.stop <- true
	}
}

// Wait for listener stops
func (c *Consumer) Wait() error {
	c.mu.Lock()
	if !c.running {
		c.mu.Unlock()
		return nil
	}

	ch := make(chan error)
	c.waitings = append(c.waitings, ch)
	c.mu.Unlock()

	return <-ch
}

func (c *Consumer) release(err error) {
	c.mu.Lock()
	for _, ch := range c.waitings {
		ch <- err
		close(ch)
	}
	c.waitings = []chan error{}
	c.mu.Unlock()
}

// GetID returns the listener ID
func (c *Consumer) GetID() string {
	return c.ID
}

// IsRunning returns consumer running status
func (c *Consumer) IsRunning() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.running
}
