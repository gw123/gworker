package amqpool

import (
	"container/list"
	"sync"
)

// ChannelPool is a pool of amqp channels
type ChannelPool struct {
	max  int
	open func(*ChannelPool) (*PooledChannel, error)

	mu       sync.Mutex
	cond     *sync.Cond
	active   int
	idles    list.List
	channels []*PooledChannel
}

// Get returns a pooled channel from the pool
func (p *ChannelPool) Get() (*PooledChannel, error) {
	p.mu.Lock()

	for {
		// Get idle channel
		for i, n := 0, p.idles.Len(); i < n; i++ {
			e := p.idles.Front()
			ic := e.Value.(*PooledChannel)
			p.idles.Remove(e)
			p.active++
			p.mu.Unlock()

			if !ic.free && ic.isConnected() {
				return ic, nil
			}

			p.mu.Lock()
			p.release()
		}

		// Create a new channel
		if p.active < p.max {
			p.active++
			p.mu.Unlock()

			c, err := p.open(p)
			if err != nil {
				p.mu.Lock()
				p.release()
				p.mu.Unlock()
				return nil, err
			}

			p.mu.Lock()
			p.channels = append(p.channels, c)
			p.mu.Unlock()

			return c, nil
		}

		if p.cond == nil {
			p.cond = sync.NewCond(&p.mu)
		}

		p.cond.Wait()
	}
}

func (p *ChannelPool) release() {
	p.active--
	if p.cond != nil {
		p.cond.Signal()
	}
}

func (p *ChannelPool) put(c *PooledChannel) {
	p.mu.Lock()

	if !c.free && c.isConnected() {
		p.idles.PushFront(c)
	}

	p.release()
	p.mu.Unlock()
}

// this is called when connection is lost
func (p *ChannelPool) clear() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.active = 0

	for _, ch := range p.channels {
		ch.setDisconnected()
	}

	p.channels = []*PooledChannel{}
}

// PooledChannel is the channel in the channel pool. It has two functions to handle with the pool.
type PooledChannel struct {
	*Channel

	free bool
	pool *ChannelPool
}

// Release put the channel back to the pool.
func (c *PooledChannel) Release() {
	if c.pool != nil {
		c.pool.put(c)
	}
}

// Free unlinks the channel with its pool
func (c *PooledChannel) Free() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.free = true
	c.Release()
	c.pool = nil
}
