package rmq

import (
	"errors"
	genericSync "github.com/aliforever/go-generic-sync-map"
	"github.com/rabbitmq/amqp091-go"
)

var (
	deliveryQueueNotSet = errors.New("delivery_queue_not_set")
)

type Consumer struct {
	ch *amqp091.Channel

	name      string
	queueName string
	autoAck   bool
	exclusive bool
	noLocal   bool
	noWait    bool
	args      map[string]interface{}
}

func NewConsumer(ch *amqp091.Channel, name, queueName string) *Consumer {
	return &Consumer{
		ch:        ch,
		name:      name,
		queueName: queueName,
		args:      map[string]interface{}{},
	}
}

func (c *Consumer) SetAutoAck() *Consumer {
	c.autoAck = true

	return c
}

func (c *Consumer) SetExclusive() *Consumer {
	c.exclusive = true

	return c
}

func (c *Consumer) SetNoLocal() *Consumer {
	c.noLocal = true

	return c
}

func (c *Consumer) SetNoWait() *Consumer {
	c.noWait = true

	return c
}

func (c *Consumer) AddArg(key string, val interface{}) *Consumer {
	c.args[key] = val

	return c
}

func (c *Consumer) Consume() (<-chan amqp091.Delivery, error) {
	if len(c.args) == 0 {
		c.args = nil
	}

	return c.ch.Consume(c.queueName, c.name, c.autoAck, c.exclusive, c.noLocal, c.noWait, c.args)
}

// ConsumeWithResponses acts as a middleware and writes to response channel if the delivery contains replyTo
// - ignoreResponses set to true will ignore delivering the events with replyTo available to the output channel
func (c *Consumer) ConsumeWithResponses(
	deliveryQueue *genericSync.Map[chan amqp091.Delivery], ignoreResponses bool) (<-chan amqp091.Delivery, error) {

	if deliveryQueue == nil {
		return nil, deliveryQueueNotSet
	}

	var outputChan = make(chan amqp091.Delivery)

	if len(c.args) == 0 {
		c.args = nil
	}

	ch, err := c.ch.Consume(c.queueName, c.name, c.autoAck, c.exclusive, c.noLocal, c.noWait, c.args)
	if err != nil {
		return nil, err
	}

	go c.deliver(ch, outputChan, deliveryQueue, ignoreResponses)

	return outputChan, nil
}

func (c *Consumer) deliver(
	source <-chan amqp091.Delivery, target chan<- amqp091.Delivery,
	deliveryQueue *genericSync.Map[chan amqp091.Delivery], ignoreResponses bool) {

	for delivery := range source {
		// fmt.Println(fmt.Sprintf("CorrelationID: %s - ReplyTo: %s", delivery.CorrelationId, delivery.ReplyTo))
		if delivery.ReplyTo != "" {
			if ch, exists := deliveryQueue.LoadAndDelete(delivery.ReplyTo); exists {
				ch <- delivery
			}
			if ignoreResponses {
				continue
			}
		}
		target <- delivery
	}
}
