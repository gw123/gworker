package amqpool

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestConnectionPoolGet(t *testing.T) {
	var expectErr error
	created := 0
	p := NewConnectionPool(
		1,
		func(ctx context.Context) (*Connection, error) {
			created++
			return &Connection{connected: true}, expectErr
		},
	)

	c, _ := p.Get(context.Background())
	if p.active != 1 {
		t.Errorf("expect active counter increased to 1, actual is %d", p.active)
	}
	if created != 1 {
		t.Errorf("expect open func is called once")
	}

	go c.Release()
	c1, _ := p.Get(context.Background())
	if c != c1 {
		t.Errorf("expect connection from idle list")
	}
	if created != 1 {
		t.Errorf("expect open func is not called this time")
	}

	c1.connected = false
	c1.close = make(chan bool)
	c1.Release()

	// test when open func returns error
	expectErr = errors.New("test error")
	go func() {
		_, err := p.Get(context.Background())
		if err != expectErr {
			t.Errorf("expect an error returned")
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// test connection Close func is called
	select {
	case <-c1.close:
		break
	case <-ctx.Done():
		t.Errorf("pool.get timeout")
	}
}

func TestConnectionPoolRelease(t *testing.T) {
	p := &ConnectionPool{}
	p.cond = sync.NewCond(&p.mu)

	done := make(chan bool)
	go func() {
		p.mu.Lock()
		go func() {
			p.mu.Lock()
			p.release()
			p.mu.Unlock()
		}()
		p.cond.Wait()
		p.mu.Unlock()
		done <- true
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	select {
	case <-done:
		if p.active != -1 {
			t.Errorf("active counter not decreased")
		}
	case <-ctx.Done():
		t.Errorf("release timeout")
	}
}

func TestConnectionPoolPut(t *testing.T) {
	p := &ConnectionPool{max: 10, active: 1}
	c := &PooledConnection{Connection: &Connection{}, pool: p}

	p.put(c)

	n := p.idles.Len()
	if n != 1 {
		t.Errorf("pool idle list length should be 1, actual is %d", n)
	}

	if p.active != 0 {
		t.Errorf("active counter is not decreased")
	}
}

func TestConnectionPoolMax(t *testing.T) {
	cases := []int{1, 2, 3, 4, 5, 6}

	for _, m := range cases {
		p := &ConnectionPool{max: m}
		if p.Max() != m {
			t.Errorf("expect connection pool Max returns %d, actual is %d", m, p.Max())
		}
	}
}

func TestPooledConnectionRelease(t *testing.T) {
	p := &ConnectionPool{max: 10}
	c := &PooledConnection{Connection: &Connection{}, pool: p}

	if p.idles.Len() > 0 {
		t.Errorf("pool idle list length should be 0")
	}

	c.Release()

	n := p.idles.Len()
	if n != 1 {
		t.Errorf("pool idle list length should be 1, actual is %d", n)
	}
}
