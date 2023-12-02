package rmq

import (
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	"time"
)

type Consumer struct {
	consumerBuilder *ConsumerBuilder

	delivery chan amqp091.Delivery

	errChan chan error
}

func newConsumer(
	consumerBuilder *ConsumerBuilder,
	retryCount int,
	retryDuration time.Duration,
) (*Consumer, error) {
	consumer := &Consumer{
		consumerBuilder: consumerBuilder,
		delivery:        make(chan amqp091.Delivery),
		errChan:         make(chan error),
	}

	tried := retryCount

	var (
		consumerChan <-chan amqp091.Delivery
		err          error
	)
	for tried > 0 {
		ch := consumerBuilder.channel.channel()

		if consumerBuilder.prefetch > 0 {
			err = ch.Qos(consumerBuilder.prefetch, 0, false)
			if err != nil {
				tried--
				time.Sleep(retryDuration)
				continue
			}
		}

		consumerChan, err = ch.Consume(
			consumerBuilder.queueName,
			consumerBuilder.name,
			consumerBuilder.autoAck,
			consumerBuilder.exclusive,
			consumerBuilder.noLocal,
			consumerBuilder.noWait,
			consumerBuilder.args,
		)
		if err != nil {
			tried--
			time.Sleep(retryDuration)
			continue
		}

		stopChan := make(chan bool)

		go consumer.process(stopChan, consumerChan)

		go consumer.keepAlive(stopChan)

		return consumer, nil
	}

	return nil, fmt.Errorf("failed to create consumer after %d retries: %s", retryCount, err)
}

func (c *Consumer) Messages() <-chan amqp091.Delivery {
	return c.delivery
}

// ErrorChan returns a channel that will receive an error when the consumer is closed
func (c *Consumer) ErrorChan() <-chan error {
	return c.errChan
}

// Cancel stops the consumer
func (c *Consumer) Cancel() error {
	return c.consumerBuilder.channel.channel().Cancel(c.consumerBuilder.name, false)
}

func (c *Consumer) process(stopChan chan bool, consumer <-chan amqp091.Delivery) {
	for delivery := range consumer {
		c.delivery <- delivery
	}

	stopChan <- true
}

// keepAlive keeps the consumer alive by recreating it when it's closed
func (c *Consumer) keepAlive(stopChan chan bool) {
	tried := c.consumerBuilder.retryCount

	var (
		consumerChan <-chan amqp091.Delivery
		err          error
	)

	for tried > 0 {
		<-stopChan

		consumerChan, err = c.consumerBuilder.channel.channel().Consume(
			c.consumerBuilder.queueName,
			c.consumerBuilder.name,
			c.consumerBuilder.autoAck,
			c.consumerBuilder.exclusive,
			c.consumerBuilder.noLocal,
			c.consumerBuilder.noWait,
			c.consumerBuilder.args,
		)
		if err != nil {
			tried--
			time.Sleep(c.consumerBuilder.retryDelay)
			continue
		}

		go c.process(stopChan, consumerChan)

		tried = c.consumerBuilder.retryCount
	}

	c.errChan <- fmt.Errorf(
		"failed to create consumer after %d retries: %s",
		c.consumerBuilder.retryCount-tried,
		err,
	)
}
