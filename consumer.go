package rmq

import (
	"errors"
	"fmt"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type ConsumerImpl interface {
	Messages() <-chan amqp091.Delivery
	ErrorChan() <-chan error
	Cancel() error
}

type Consumer struct {
	consumerBuilder *ConsumerBuilder

	delivery chan amqp091.Delivery

	errChan chan error

	consumerClosedChan chan error
}

func newConsumer(
	consumerBuilder *ConsumerBuilder,
	retryCount int,
	retryDuration time.Duration,
) (*Consumer, error) {
	consumer := &Consumer{
		consumerBuilder:    consumerBuilder,
		delivery:           make(chan amqp091.Delivery),
		errChan:            make(chan error),
		consumerClosedChan: make(chan error),
	}

	consumerBuilder.channel.addCloseChannel(consumer.consumerClosedChan)

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

		go consumer.process(consumerChan)

		go consumer.keepAlive()

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

func (c *Consumer) process(consumer <-chan amqp091.Delivery) {
	for delivery := range consumer {
		c.delivery <- delivery
	}
}

// keepAlive keeps the consumer alive by recreating it when it's closed
func (c *Consumer) keepAlive() {
	defer close(c.delivery)

	stopChan := make(chan bool)

	tried := c.consumerBuilder.retryCount

	var (
		consumerChan <-chan amqp091.Delivery
		err          error
	)

	for tried > 0 {
		select {
		case <-stopChan:
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
				fmt.Println("failed to create consumer 1", err)
				tried--
				time.Sleep(c.consumerBuilder.retryDelay)
				continue
			}

			go c.process(consumerChan)

			tried = c.consumerBuilder.retryCount
		case closeErr := <-c.consumerClosedChan:
			if closeErr == nil {
				c.errChan <- errors.New("failed to create consumer due to channel failure")
				return
			}

			go func() {
				stopChan <- true
			}()
		}
	}

	c.errChan <- fmt.Errorf(
		"failed to create consumer after %d retries: %s",
		c.consumerBuilder.retryCount-tried,
		err,
	)
}
