package rmq

import "github.com/rabbitmq/amqp091-go"

type Delivery struct {
	*amqp091.Delivery

	ch *amqp091.Channel
}

func newDelivery(ch *amqp091.Channel, delivery amqp091.Delivery) *Delivery {
	return &Delivery{
		Delivery: &delivery,
		ch:       ch,
	}
}

func (d *Delivery) Ack() error {
	return d.Ack()
}

func (d *Delivery) NAck(multiple, requeue bool) error {
	return d.Nack(multiple, requeue)
}
