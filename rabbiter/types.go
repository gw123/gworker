package rabbiter

import "github.com/streadway/amqp"

const (
	// HeaderAttempt is field name of message trying times number.
	HeaderAttempt = "Rabbiter-Attempt"
)

// Queue struct is queue properties for AMQP.
// RoutingKey and PrefetchCount is not AMQP Queue properties,
// they are used by publisher and consumer for convenience.
type Queue struct {
	Name        string
	Durable     bool
	AutoDeleted bool
	Exclusive   bool
	NoWait      bool
	Args        amqp.Table

	RoutingKey    string
	PrefetchCount int
}

// Exchange struct is exchange properties for AMQP.
type Exchange struct {
	Name        string
	Type        string
	Durable     bool
	AutoDeleted bool
	Internal    bool
	NoWait      bool
	Args        amqp.Table
}

// Message struct contains basic message data.
type Message struct {
	ID          string
	Body        []byte
	ContentType string
	RoutingKey  string
	Headers     amqp.Table
	Attempt     int
}

type Factory interface {
	Publisher() Publisher
	Consumer(*Exchange, *Queue, HandlerFunc) Consumer
}
