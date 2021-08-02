package amqpool

import (
	"container/list"
	"context"
	"sync"
)

// ConnectionOpener is a function used by pool to get a Connection
type ConnectionOpener func(context.Context) (*Connection, error)

// ConnectionPool is pool of Connection
type ConnectionPool struct {
	max  int
	open ConnectionOpener

	mu     sync.Mutex
	cond   *sync.Cond
	active int
	idles  list.List
}

// NewConnectionPool returns a new connection pool
func NewConnectionPool(max int, open ConnectionOpener) *ConnectionPool {
	return &ConnectionPool{
		max:  max,
		open: open,
	}
}

// Max returns the max connection number
func (p *ConnectionPool) Max() int {
	return p.max
}

// Get returns a pooled connection from the pool
func (p *ConnectionPool) Get(ctx context.Context) (*PooledConnection, error) {
	p.mu.Lock()

	for {
		for i, n := 0, p.idles.Len(); i < n; i++ {
			e := p.idles.Front()
			pc := e.Value.(*PooledConnection)
			p.idles.Remove(e)
			p.active++
			p.mu.Unlock()

			if pc.IsConnected() {
				return pc, nil
			}

			pc.Close()

			p.mu.Lock()
			p.release()
		}

		if p.active < p.max {
			p.active++
			p.mu.Unlock()

			c, err := p.open(ctx)
			if err != nil {
				p.mu.Lock()
				p.release()
				p.mu.Unlock()
				return nil, err
			}

			pc := &PooledConnection{
				Connection: c,
				pool:       p,
			}
			return pc, nil
		}

		if p.cond == nil {
			p.cond = sync.NewCond(&p.mu)
		}

		p.cond.Wait()
	}
}

func (p *ConnectionPool) release() {
	p.active--
	if p.cond != nil {
		p.cond.Signal()
	}
}

func (p *ConnectionPool) put(c *PooledConnection) {
	p.mu.Lock()

	p.idles.PushFront(c)
	p.release()

	p.mu.Unlock()
}

// PooledConnection is wrapper of connection
type PooledConnection struct {
	*Connection

	pool *ConnectionPool
}

// Release put connection back to the pool
func (c *PooledConnection) Release() {
	c.pool.put(c)
}
