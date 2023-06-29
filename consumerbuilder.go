package rmq

import (
	"errors"
	"time"
)

var (
	deliveryQueueNotSet = errors.New("delivery_queue_not_set")
)

type ConsumerBuilder struct {
	channel *Channel

	name      string
	queueName string
	autoAck   bool
	exclusive bool
	noLocal   bool
	noWait    bool
	args      map[string]interface{}

	retryCount int
	retryDelay time.Duration
}

func newConsumerBuilder(
	channel *Channel,
	retryCount int,
	retryDelay time.Duration,
	name string,
	queueName string,
) *ConsumerBuilder {
	return &ConsumerBuilder{
		channel:    channel,
		name:       name,
		queueName:  queueName,
		args:       map[string]interface{}{},
		retryCount: retryCount,
		retryDelay: retryDelay,
	}
}

func (c *ConsumerBuilder) SetAutoAck() *ConsumerBuilder {
	c.autoAck = true

	return c
}

func (c *ConsumerBuilder) SetExclusive() *ConsumerBuilder {
	c.exclusive = true

	return c
}

func (c *ConsumerBuilder) SetNoLocal() *ConsumerBuilder {
	c.noLocal = true

	return c
}

func (c *ConsumerBuilder) SetNoWait() *ConsumerBuilder {
	c.noWait = true

	return c
}

func (c *ConsumerBuilder) AddArg(key string, val interface{}) *ConsumerBuilder {
	c.args[key] = val

	return c
}

func (c *ConsumerBuilder) Build() (*Consumer, error) {
	return newConsumer(c, c.retryCount, c.retryDelay)
}

// ConsumeWithResponses acts as a middleware and writes to response channel if the delivery contains replyTo
// - ignoreResponses set to true will ignore delivering the events with replyTo available to the output channel
// func (c *consumerBuilder) ConsumeWithResponses(
// 	deliveryQueue *genericSync.Map[chan amqp091.Delivery], ignoreResponses bool) (<-chan amqp091.Delivery, error) {
//
// 	if deliveryQueue == nil {
// 		return nil, deliveryQueueNotSet
// 	}
//
// 	var outputChan = make(chan amqp091.Delivery)
//
// 	if len(c.args) == 0 {
// 		c.args = nil
// 	}
//
// 	ch, err := c.ch.Consume(c.queueName, c.name, c.autoAck, c.exclusive, c.noLocal, c.noWait, c.args)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	go c.deliver(ch, outputChan, deliveryQueue, ignoreResponses)
//
// 	return outputChan, nil
// }
//
//
// func (c *consumerBuilder) deliver(
// 	source <-chan amqp091.Delivery, target chan<- amqp091.Delivery,
// 	deliveryQueue *genericSync.Map[chan amqp091.Delivery], ignoreResponses bool) {
//
// 	for delivery := range source {
// 		// fmt.Println(fmt.Sprintf("CorrelationID: %s - ReplyTo: %s", delivery.CorrelationId, delivery.ReplyTo))
// 		if delivery.ReplyTo != "" {
// 			if ch, exists := deliveryQueue.LoadAndDelete(delivery.ReplyTo); exists {
// 				ch <- delivery
// 			}
// 			if ignoreResponses {
// 				continue
// 			}
// 		}
// 		target <- delivery
// 	}
// }
