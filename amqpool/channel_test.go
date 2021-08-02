package amqpool

import (
	"sync"
	"testing"

	"github.com/streadway/amqp"
)

type FakeChannel struct {
	delivery        chan amqp.Delivery
	consumerTag     string
	consumeQueue    string
	closed          bool
	queue           string
	queueBinding    [3]string
	exchange        string
	exchangeBinding [3]string
	prefetch        int

	mux                sync.RWMutex
	err                error
	consumeErr         error
	closeErr           error
	declareExchangeErr error
	bindExchangeErr    error
	publishErr         error
	declareQueueErr    error
	bindQueueErr       error
	unbindQueueErr     error
	qosErr             error
}

func NewFakeChannel() *FakeChannel {
	c := &FakeChannel{
		delivery: make(chan amqp.Delivery),
		closed:   false,
	}
	return c
}

func (c *FakeChannel) Consume(queue string, consumer string, autoAck bool, exclusive bool, noLocal bool, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	c.consumeQueue = queue
	c.consumerTag = consumer
	return c.delivery, c.err
}

func (c *FakeChannel) Close() error {
	c.mux.RLock()
	defer c.mux.RUnlock()
	c.closed = true
	if c.closeErr != nil {
		return c.closeErr
	}
	return c.err
}

func (c *FakeChannel) ExchangeBind(destination string, key string, source string, noWait bool, args amqp.Table) error {
	c.exchangeBinding[0] = destination
	c.exchangeBinding[1] = key
	c.exchangeBinding[2] = source
	if c.bindExchangeErr != nil {
		return c.bindExchangeErr
	}
	return c.err
}

func (c *FakeChannel) ExchangeDeclare(name string, kind string, durable bool, autoDelete bool, internal bool, noWait bool, args amqp.Table) error {
	c.exchange = name
	if c.declareExchangeErr != nil {
		return c.declareExchangeErr
	}
	return c.err
}

func (c *FakeChannel) Publish(exchange string, key string, mandatory bool, immediate bool, msg amqp.Publishing) error {
	if c.publishErr != nil {
		return c.publishErr
	}
	return c.err
}

func (c *FakeChannel) QueueBind(name string, key string, exchange string, noWait bool, args amqp.Table) error {
	c.queueBinding[0] = name
	c.queueBinding[1] = key
	c.queueBinding[2] = exchange
	if c.bindQueueErr != nil {
		return c.bindQueueErr
	}
	return c.err
}

func (c *FakeChannel) QueueUnbind(name string, key string, exchange string, args amqp.Table) error {
	if c.unbindQueueErr != nil {
		return c.unbindQueueErr
	}
	return c.err
}

func (c *FakeChannel) QueueDeclare(name string, durable bool, autoDelete bool, exclusive bool, noWait bool, args amqp.Table) (amqp.Queue, error) {
	c.queue = name
	q := amqp.Queue{}
	if c.declareQueueErr != nil {
		return q, c.declareQueueErr
	}
	return q, c.err
}

func (c *FakeChannel) Qos(prefetchCount, prefetchSize int, global bool) error {
	c.prefetch = prefetchCount
	if c.qosErr != nil {
		return c.qosErr
	}
	return c.err
}

func TestSetDisconnected(t *testing.T) {
	ch := &Channel{connected: true}
	ch.setDisconnected()
	if ch.connected {
		t.Errorf("expect connected status is false")
	}
}

func TestChannelIsConnected(t *testing.T) {
	ch := &Channel{}

	if ch.isConnected() {
		t.Errorf("expect connected status is false")
	}

	ch.connected = true
	if !ch.isConnected() {
		t.Errorf("expect connected status is true")
	}
}
