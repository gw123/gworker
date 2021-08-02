package amqpool

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/streadway/amqp"
)

func TestChannelPoolGet(t *testing.T) {
	created := 0
	p := &ChannelPool{
		max: 1,
		open: func(p *ChannelPool) (*PooledChannel, error) {
			created++
			return &PooledChannel{
				Channel: &Channel{connected: true},
				pool:    p,
			}, nil
		},
	}

	c, err := p.Get()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	// release an active channel, function get will return an idle channel
	// from idle channel list
	go c.Release()
	c1, _ := p.Get()

	if p.active != 1 {
		t.Errorf("expect active channel number is 1, actual is %d", p.active)
	}

	if p.idles.Len() != 0 {
		t.Errorf("pool idle list is not empty")
	}

	if created != 1 {
		t.Errorf("expect created is %d, actual is %d", 1, created)
	}

	if c != c1 {
		t.Errorf("expect an old channel from idle list")
	}

	// channel is disconnected, pool will create a new channel
	c1.connected = false
	go func() { c1.Release() }()
	c, _ = p.Get()

	if created != 2 {
		t.Errorf("expect created is %d, actual is %d", 2, created)
	}

	if c == c1 {
		t.Errorf("expect a new pooled channel")
	}

	if p.active != 1 {
		t.Errorf("expect active channel number is 1, actual is %d", p.active)
	}

	// test disconnected channel in idle list
	c.Release()
	c.setDisconnected()
	c1, _ = p.Get()

	if created != 3 {
		t.Errorf("expect created is %d, actual is %d", 3, created)
	}

	if c == c1 {
		t.Errorf("expect a new pooled channel")
	}
}

func TestChannelPoolGetFailed(t *testing.T) {
	expectError := errors.New("channel pool error")
	p := &ChannelPool{
		max: 1,
		open: func(p *ChannelPool) (*PooledChannel, error) {
			return nil, expectError
		},
	}

	_, err := p.Get()
	if err != expectError {
		t.Errorf("expected error is %s, actual is %s", expectError, err)
	}
}

func TestChannelPoolPut(t *testing.T) {
	p := &ChannelPool{}
	pc := &PooledChannel{
		Channel: &Channel{connected: true},
		pool:    p,
	}

	p.put(pc)
	if p.idles.Len() != 1 {
		t.Errorf("pooled channel not returned to idle list")
	}

	if p.active != -1 {
		t.Errorf("active counter not decreased")
	}

	pc = &PooledChannel{
		Channel: &Channel{connected: false},
		pool:    p,
	}
	p.put(pc)
	if p.idles.Len() != 1 {
		t.Errorf("returned a disconnected channel to idle list")
	}

	if p.active != -2 {
		t.Errorf("active counter not decreased twice")
	}
}

func TestChannelPoolRelease(t *testing.T) {
	p := &ChannelPool{}

	p.release()
	if p.active != -1 {
		t.Errorf("active counter not decreased")
	}

	done := make(chan bool)
	p.cond = sync.NewCond(&p.mu)
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
		break
	case <-ctx.Done():
		t.Errorf("release timeout")
	}
}

func TestChannelPoolClear(t *testing.T) {
	n := 2
	p := &ChannelPool{active: n}
	newPooledChannel := func() *PooledChannel {
		return &PooledChannel{
			Channel: &Channel{
				AMQPChannel: &amqp.Channel{},
				connected:   true,
			},
			pool: p,
		}
	}
	chs := []*PooledChannel{}
	for i := 0; i < n; i++ {
		chs = append(chs, newPooledChannel())
	}
	p.channels = chs

	p.clear()

	if p.active != 0 {
		t.Errorf("active counter is great than 0: %d", p.active)
	}

	if len(p.channels) > 0 {
		t.Errorf("opened channel list is not cleared")
	}
}

func TestPooledChannelRelease(t *testing.T) {
	p := &ChannelPool{}
	pc := &PooledChannel{
		Channel: &Channel{},
		pool:    p,
	}
	pc.connected = true

	pc.Release()
	if p.idles.Len() != 1 {
		t.Errorf("pooled channel not returned to idle list")
	}

	if p.active != -1 {
		t.Errorf("active counter not decreased")
	}

	pc.connected = false
	pc.Release()
	if p.idles.Len() != 1 {
		t.Errorf("returned a disconnected channel to idle list")
	}

	if p.active != -2 {
		t.Errorf("active counter not decreased twice")
	}
}

func TestPooledChannelFree(t *testing.T) {
	p := &ChannelPool{}
	pc := &PooledChannel{
		Channel: &Channel{},
		pool:    p,
	}
	pc.connected = true

	pc.Free()

	if !pc.free {
		t.Errorf("expect pooled channel be marked as free")
	}

	if pc.pool != nil {
		t.Errorf("expect pooled channel's pool is nil")
	}

	if p.idles.Len() != 0 {
		t.Errorf("expect pooled channel not returned to idle list")
	}

	if p.active != -1 {
		t.Errorf("expect active counter decreased")
	}
}
