package rmq

import "github.com/rabbitmq/amqp091-go"

type DeliveryImpl interface {
	Ack(multiple bool) error
	Nack(multiple, requeue bool) error
	Reject(requeue bool) error
}

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

func (d *Delivery) Ack(multiple bool) error {
	return d.Delivery.Ack(multiple)
}

func (d *Delivery) Nack(multiple, requeue bool) error {
	return d.Delivery.Nack(multiple, requeue)
}

func (d *Delivery) Reject(requeue bool) error {
	return d.Delivery.Reject(requeue)
}
