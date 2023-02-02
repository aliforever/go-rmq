package rmq

import "github.com/rabbitmq/amqp091-go"

type Queue struct {
	ch *amqp091.Channel
	*amqp091.Queue
}

func newQueue(ch *amqp091.Channel, queue *amqp091.Queue) *Queue {
	return &Queue{ch: ch, Queue: queue}
}

func (q *Queue) BindToExchange(exchange string, routingKey string, noWait bool, args map[string]interface{}) error {
	return q.ch.QueueBind(q.Name, routingKey, exchange, noWait, args)
}
