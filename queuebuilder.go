package rmq

import (
	"time"
)

type QueueBuilderImpl interface {
	SetName(name string) QueueBuilderImpl
	SetDurable() QueueBuilderImpl
	SetAutoDelete() QueueBuilderImpl
	SetExclusive() QueueBuilderImpl
	SetNoWait() QueueBuilderImpl
	AddArg(key string, val interface{}) QueueBuilderImpl
	Declare() (QueueImpl, error)
}

type QueueBuilder struct {
	channel      *Channel
	exchangeName string

	name       string
	durable    bool
	autoDelete bool
	exclusive  bool
	noWait     bool
	args       map[string]interface{}

	retryCount int
	retryDelay time.Duration
}

func newQueueBuilder(channel *Channel, retryCount int, retryDelay time.Duration) *QueueBuilder {
	return &QueueBuilder{
		channel:    channel,
		args:       map[string]interface{}{},
		retryCount: retryCount,
		retryDelay: retryDelay,
	}
}

func (q *QueueBuilder) SetName(name string) *QueueBuilder {
	q.name = name

	return q
}

func (q *QueueBuilder) SetDurable() *QueueBuilder {
	q.durable = true

	return q
}

func (q *QueueBuilder) SetAutoDelete() *QueueBuilder {
	q.autoDelete = true

	return q
}

func (q *QueueBuilder) SetExclusive() *QueueBuilder {
	q.exclusive = true

	return q
}

func (q *QueueBuilder) SetNoWait() *QueueBuilder {
	q.noWait = true

	return q
}

func (q *QueueBuilder) AddArg(key string, val interface{}) *QueueBuilder {
	q.args[key] = val

	return q
}

func (q *QueueBuilder) Declare() (*Queue, error) {
	return newQueue(q)
}
