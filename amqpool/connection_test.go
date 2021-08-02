package amqpool

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/streadway/amqp"
)

type FakeAmqpConnection struct {
	receiver chan *amqp.Error
	err      error
}

func (c *FakeAmqpConnection) Channel() (*amqp.Channel, error) {
	return &amqp.Channel{}, c.err
}

func (c *FakeAmqpConnection) Close() error {
	return c.err
}

func (c *FakeAmqpConnection) NotifyClose(receiver chan *amqp.Error) chan *amqp.Error {
	c.receiver = receiver
	return receiver
}

func TestNewConnection(t *testing.T) {
	options := ConnectionOptions{
		Dial:          func() (AMQPConnection, error) { return &FakeAmqpConnection{}, nil },
		Logger:        NopLogger,
		RetryInterval: 1 * time.Second,
	}
	c := NewConnection(options)

	if c.Channels == nil {
		t.Errorf("channel pool is not inited")
	}

	if c.Channels.max != DefaultChannels {
		t.Errorf("pool max channel number is incorrect, expect %d, actual %d", DefaultChannels, c.Channels.max)
	}
}

func TestConnect(t *testing.T) {
	fa := &FakeAmqpConnection{}
	c := NewConnection(ConnectionOptions{
		Logger:        NopLogger,
		RetryInterval: 1 * time.Second,
		Dial: func() (AMQPConnection, error) {
			return fa, nil
		},
	})

	c.connected = true
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	done := make(chan bool)
	defer close(done)

	go func() {
		c.Connect(context.Background())
		done <- true
	}()

	select {
	case <-done:
		if c.Amqp != nil {
			t.Errorf("should not call dial func")
		}
	case <-ctx.Done():
		t.Errorf("connect timeout")
	}

	c.connected = false
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	done = make(chan bool)
	defer close(done)

	go func() {
		c.Connect(context.Background())
		done <- true
	}()

	select {
	case <-done:
		if c.Amqp == nil {
			t.Errorf("expect amqp not nil")
		}
		if c.connected == false {
			t.Errorf("expect connected status is true")
		}
		if fa.receiver == nil {
			t.Errorf("expect NotifyClose is called")
		}
	case <-ctx.Done():
		t.Errorf("connect timeout")
	}

	close(fa.receiver)
}

func TestConnectionConnectWithError(t *testing.T) {
	n := 0
	dialError := errors.New("fake dial error")
	fa := &FakeAmqpConnection{}

	c := NewConnection(ConnectionOptions{
		Logger:        NopLogger,
		RetryInterval: 1 * time.Second,
		Dial: func() (AMQPConnection, error) {
			if n == 0 {
				n++
				return nil, dialError
			}
			return fa, nil
		},
	})
	c.sleep = 5 * time.Millisecond

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	done := make(chan bool)
	defer close(done)

	go func() {
		c.Connect(context.Background())
		done <- true
	}()

	select {
	case <-done:
		if !c.connected {
			t.Errorf("expect connection is connected")
		}
		close(fa.receiver)
	case <-ctx.Done():
		t.Errorf("connect timeout")
	}

	n = 0
	c.mux.Lock()
	c.connected = false
	c.Amqp = nil
	c.mux.Unlock()

	edone := make(chan error)
	defer close(edone)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Millisecond)
		defer cancel()
		edone <- c.Connect(ctx)
	}()

	select {
	case err := <-edone:
		if err == ErrConnectionStopped {
			t.Errorf("expect timeout error")
		}
		break
	case <-ctx.Done():
		t.Errorf("connect op timeout")
	}

	go func() {
		n = 0
		go c.Close()
		edone <- c.Connect(context.Background())
	}()

	select {
	case err := <-edone:
		if err != ErrConnectionStopped {
			t.Errorf("expect connection stop error, actual is %s", err)
		}
		break
	case <-ctx.Done():
		t.Errorf("connect op timeout")
	}
}

func TestConnectionWatchAndReconnect(t *testing.T) {
	notify := make(chan *amqp.Error)
	defer close(notify)
	fa := &FakeAmqpConnection{}

	c := NewConnection(ConnectionOptions{
		Logger:        NopLogger,
		RetryInterval: 1 * time.Second,
		Dial: func() (AMQPConnection, error) {
			return fa, nil
		},
	})
	c.connected = true
	c.Channels.active = 1
	c.Amqp = fa

	go func() {
		notify <- amqp.ErrClosed
	}()

	done := make(chan bool)
	defer close(done)

	go func() {
		c.watchAndReconnect(notify)
		done <- true
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	select {
	case <-done:
		if c.Amqp == nil {
			t.Errorf("expect reconnecting performed")
		}
		if c.Channels.active > 0 {
			t.Errorf("expect channel pool is cleared")
		}
	case <-ctx.Done():
		t.Errorf("reconnect timeout")
	}
}

func TestConnectionClose(t *testing.T) {
	fa := &FakeAmqpConnection{}

	c := NewConnection(ConnectionOptions{
		Logger:        NopLogger,
		RetryInterval: 1 * time.Second,
		Dial: func() (AMQPConnection, error) {
			return fa, nil
		},
	})

	done := make(chan bool)
	defer close(done)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	go func() {
		c.Connect(ctx)
		c.Close()
		done <- true
	}()

	select {
	case <-done:
		c.mux.RLock()
		if c.Amqp != nil {
			t.Errorf("expect amqp is nil")
		}
		c.mux.RUnlock()
	case <-ctx.Done():
		t.Errorf("close timeout")
	}
}

func TestConnectionIsConnected(t *testing.T) {
	cases := [2]bool{true, false}
	for _, status := range cases {
		c := &Connection{connected: status}
		actual := c.IsConnected()
		if actual != status {
			t.Errorf("expect connection status is %v, actual is %v", status, actual)
		}
	}
}

func TestConnectionPooledChannel(t *testing.T) {
	fa := &FakeAmqpConnection{}
	c := NewConnection(ConnectionOptions{
		Logger:        NopLogger,
		RetryInterval: 1 * time.Second,
	})
	c.Amqp = fa

	ch, err := c.ChannelWithPool(c.Channels)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if ch.pool != c.Channels {
		t.Errorf("expect pool is %v, actual is %v", c.Channels, ch.pool)
	}

	if !ch.connected {
		t.Errorf("pooled channel connection status error")
	}

	fa.err = errors.New("fake error")
	ch, err = c.ChannelWithPool(c.Channels)

	if err != fa.err {
		t.Errorf("expect error is %v, actual is %v", fa.err, err)
	}
}

func TestConnectionChannel(t *testing.T) {
	fa := &FakeAmqpConnection{}
	c := NewConnection(ConnectionOptions{
		Logger:        NopLogger,
		RetryInterval: 1 * time.Second,
		MaxChannels:   10,
	})
	c.Amqp = fa

	n := 5
	var ch PooledAMQPChannel
	for i := 0; i < n; i++ {
		ch, _ = c.Channel()
	}

	pooledChannel := ch.(*PooledChannel)
	if pooledChannel.pool != c.Channels {
		t.Errorf("expect pool is %v, actual is %v", c.Channels, pooledChannel.pool)
	}

	if !pooledChannel.connected {
		t.Errorf("pooled channel connection status error")
	}

	if c.Channels.active != n {
		t.Errorf("channel is not from pool")
	}
}
