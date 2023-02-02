package rmq

import "github.com/rabbitmq/amqp091-go"

type DeliveryChannel chan amqp091.Delivery
