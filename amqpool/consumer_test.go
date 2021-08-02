package amqpool

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/streadway/amqp"
)

func TestConsumerListen(t *testing.T) {
	queue := "queue_name"
	channel := NewFakeChannel()

	var expectErr error

	c := &Consumer{
		ID:     "A100001",
		Logger: NopLogger,
		Open: func(ctx context.Context, pattern string) (AMQPChannel, string, error) {
			return channel, queue, expectErr
		},
		stop: make(chan bool),
	}

	tag := c.ID

	c.Listen(context.Background())

	if channel.consumerTag != tag {
		t.Errorf("expect consumer tag is %s, actual is %s", tag, channel.consumerTag)
	}

	if channel.consumeQueue != queue {
		t.Errorf("expect consumer queue is %s, actual is %s", queue, channel.consumeQueue)
	}

	if c.running != true {
		t.Errorf("expect running status is true")
	}

	err := c.Listen(context.Background())
	if err != nil {
		t.Errorf("expect listen a running listener returning nil")
	}

	c.stop <- true

	c.mu.Lock()
	if c.running {
		t.Errorf("expect stopped listener running status is false")
	}
	c.mu.Unlock()

	expectErr = errors.New("test error")
	err = c.Listen(context.Background())
	if err != expectErr {
		t.Errorf("expect listen returns error: %s, actual is %s", expectErr, err)
	}

	expectErr = nil
	channel.mux.Lock()
	channel.err = errors.New("test channel error")
	err = c.Listen(context.Background())
	if err != channel.err {
		t.Errorf("expect listen returns error: %s, actual is %s", channel.err, err)
	}
	channel.mux.Unlock()
}

func TestConsumerDo(t *testing.T) {
	channel := NewFakeChannel()

	c := &Consumer{
		ID:      "A100002",
		Logger:  NopLogger,
		running: true,
		stop:    make(chan bool),
	}

	go c.do(channel, channel.delivery)
	go c.Stop()

	err := c.Wait()
	if err != ErrConsumerStopped {
		t.Errorf("expect err is %s, actual is %s", ErrConsumerStopped, err)
	}

	if c.running != false {
		t.Errorf("expect running status is false, actual is true")
	}

	if !channel.closed {
		t.Errorf("expect channel is closed")
	}
}

func TestConsumerDoWithChannelClose(t *testing.T) {
	queue := "queue_name"
	channel := NewFakeChannel()

	c := &Consumer{
		ID:     "A100003",
		Logger: NopLogger,
		Open: func(ctx context.Context, pattern string) (AMQPChannel, string, error) {
			channel.delivery = make(chan amqp.Delivery)
			return channel, queue, nil
		},
		running: true,
		stop:    make(chan bool),
	}

	go c.do(channel, channel.delivery)

	// test channel close
	go close(channel.delivery)

	err := c.Wait()
	if err != ErrConsumerClosed {
		t.Errorf("expect err is %s, actual is %s", ErrConsumerClosed, err)
	}

	// test auto reconnect
	for {
		c.mu.RLock()
		if c.running {
			c.mu.RUnlock()
			c.Stop()
			break
		}
		c.mu.RUnlock()
	}
}

func TestConsumerHandleWithDo(t *testing.T) {
	queue := "queue_name"
	channel := NewFakeChannel()

	res := make(chan amqp.Delivery)

	c := &Consumer{
		ID:     "A100005",
		Logger: NopLogger,
		Open: func(ctx context.Context, pattern string) (AMQPChannel, string, error) {
			channel.delivery = make(chan amqp.Delivery)
			return channel, queue, nil
		},
		HandlerFunc: func(d *amqp.Delivery) error {
			res <- *d
			return nil
		},
		running: true,
		stop:    make(chan bool),
	}

	go c.do(channel, channel.delivery)
	defer c.Stop()

	cases := [][]byte{
		[]byte(`{"name":"amqpool"}`),
		[]byte("invalid protobuf"),
		[]byte("任意字符"),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	for _, c := range cases {
		go func() {
			d := amqp.Delivery{Body: c}
			channel.delivery <- d
		}()

		select {
		case d := <-res:
			// should only receive this valid event
			if !reflect.DeepEqual(d.Body, c) {
				t.Errorf("expect content is %#v, actual is %#v", c, d.Body)
			}
		case <-ctx.Done():
			t.Errorf("handle timeout")
		}
	}
}

func TestConsumerGetID(t *testing.T) {
	cases := []string{
		"A100001",
		"A100002",
		"A100003",
		"A100004",
		"A100005",
	}

	for _, expect := range cases {
		c := &Consumer{ID: expect}
		actual := c.GetID()
		if actual != expect {
			t.Errorf("expect ID is %s, actual is %s", expect, actual)
		}
	}
}

func TestConsumerStop(t *testing.T) {
	c := &Consumer{running: true}
	c.stop = make(chan bool)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	done := make(chan bool)
	defer close(done)

	go c.Stop()

	select {
	case <-c.stop:
		break
	case <-ctx.Done():
		t.Errorf("stop timeout")
	}
}

func TestConsumerWait(t *testing.T) {
	c := &Consumer{running: true}
	waitErr := errors.New("test waiting")

	go func() {
		c.mu.RLock()
		defer c.mu.RUnlock()
		expect := 1
		actual := len(c.waitings)
		if actual != expect {
			t.Errorf("Listener.Wait expect waitings length is %d, actual is %d", expect, actual)
		}
		c.waitings[0] <- waitErr
	}()

	err := c.Wait()
	if err != waitErr {
		t.Errorf("expect error is %s, actual is %s", waitErr, err)
	}
}

func TestConsumerReleaseWait(t *testing.T) {
	c := &Consumer{running: true}
	waitErr := errors.New("test waiting")
	n := 10

	wg := new(sync.WaitGroup)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			err := c.Wait()
			wg.Done()
			if err != waitErr {
				t.Errorf("expect error is %s, actual is %s", waitErr, err)
			}
		}()
	}

	done := make(chan bool)
	defer close(done)

	go func() {
		wg.Wait()
		done <- true
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	for {
		c.mu.RLock()
		if len(c.waitings) == n {
			c.mu.RUnlock()
			c.release(waitErr)
			break
		}
		c.mu.RUnlock()
	}

	select {
	case <-done:
		break
	case <-ctx.Done():
		t.Errorf("release wait timeout")
	}
}

func TestConsumerWaitStopped(t *testing.T) {
	c := &Consumer{running: false}

	done := make(chan error)
	defer close(done)
	go func() {
		err := c.Wait()
		done <- err
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("expect error is nil, actual is %s", err)
		}
	case <-ctx.Done():
		t.Errorf("wait stopped timeout")
	}
}
