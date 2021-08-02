package amqpool

import (
	"sync"
)

// Channel is AmqpChannel wrapper used by this package.
type Channel struct {
	AMQPChannel

	mu        sync.RWMutex
	connected bool
}

func (c *Channel) setDisconnected() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.connected = false
}

func (c *Channel) isConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.connected
}
