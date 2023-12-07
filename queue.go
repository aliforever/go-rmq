package rmq

import (
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	"time"
)

type QueueImpl interface {
	Name() string
	BindToExchange(
		exchange string,
		routingKey string,
		noWait bool,
		args map[string]interface{},
	) error
}

type Queue struct {
	queueBuilder *QueueBuilder

	queue amqp091.Queue
}

func newQueue(builder *QueueBuilder) (*Queue, error) {
	qb := &Queue{queueBuilder: builder}

	leftTries := builder.retryCount

	var err error

	for leftTries > 0 {
		qb.queue, err = builder.channel.channel().QueueDeclare(
			builder.name,
			builder.durable,
			builder.autoDelete,
			builder.exclusive,
			builder.noWait,
			builder.args,
		)
		if err != nil {
			leftTries--
			time.Sleep(builder.retryDelay)
			continue
		}

		return qb, nil
	}

	return nil, fmt.Errorf("failed to declare queue after %d tries: %s", builder.retryCount, err)
}

func (q *Queue) Name() string {
	return q.queue.Name
}

func (q *Queue) BindToExchange(
	exchange string,
	routingKey string,
	noWait bool,
	args map[string]interface{},
) error {
	leftTries := q.queueBuilder.retryCount

	var err error

	for leftTries > 0 {
		err = q.queueBuilder.channel.channel().QueueBind(
			q.queueBuilder.name,
			routingKey,
			exchange,
			noWait,
			args,
		)
		if err != nil {
			leftTries--
			time.Sleep(q.queueBuilder.retryDelay)
			continue
		}

		return nil
	}

	return fmt.Errorf("failed to bind queue to exchange after %d tries: %s", q.queueBuilder.retryCount, err)
}
