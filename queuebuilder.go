package rmq

import (
	"github.com/rabbitmq/amqp091-go"
)

type QueueBuilder struct {
	conn         *amqp091.Connection
	ch           *amqp091.Channel
	exchangeName string

	name       string
	durable    bool
	autoDelete bool
	exclusive  bool
	noWait     bool
	args       map[string]interface{}
}

func NewQueueBuilder(ch *amqp091.Channel) *QueueBuilder {
	return &QueueBuilder{
		ch:   ch,
		args: map[string]interface{}{},
	}
}

func NewQueueBuilderWithChannel(conn *amqp091.Connection) *QueueBuilder {
	return &QueueBuilder{
		conn: conn,
		args: map[string]interface{}{},
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
	if len(q.args) == 0 {
		q.args = nil
	}

	ch, err := q.getChannel()
	if err != nil {
		return nil, err
	}

	queue, err := ch.QueueDeclare(q.name, q.durable, q.autoDelete, q.exclusive, q.noWait, q.args)
	if err != nil {
		return nil, err
	}

	return newQueue(ch, &queue), nil
}

func (q *QueueBuilder) getChannel() (*amqp091.Channel, error) {
	if q.ch != nil {
		return q.ch, nil
	}

	if q.conn == nil {
		return nil, ConnectionNotSetError
	}

	return q.conn.Channel()
}
