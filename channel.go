package rmq

import (
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	"sync"
	"time"
)

type ChannelImpl interface {
	PublisherBuilder(exchange string, routingKey string) PublisherBuilderImpl
	ConsumerBuilder(name, queue string) ConsumerBuilderImpl
	QueueBuilder() QueueBuilderImpl
	ExchangeBuilder(name string) ExchangeBuilderImpl
	FanoutExchangeBuilder(name string) ExchangeBuilderImpl
	DirectExchangeBuilder(name string) ExchangeBuilderImpl
	TopicExchangeBuilder(name string) ExchangeBuilderImpl
	CloseChan() <-chan error
}

type Channel struct {
	m sync.Mutex

	rmq *RMQ

	ch *amqp091.Channel

	retryTimes int
	retryDelay time.Duration

	closeChan chan error
}

func newChannel(rmq *RMQ, retryTimes int, retryDelay time.Duration, withConfirm bool) (*Channel, error) {
	timesTried := retryTimes

	c := &Channel{
		rmq:        rmq,
		closeChan:  make(chan error),
		retryTimes: retryTimes,
		retryDelay: retryDelay,
	}

	var (
		ch  *amqp091.Channel
		err error
	)

	for timesTried > 0 {
		ch, err = rmq.conn.Channel()
		if err != nil {
			timesTried--
			time.Sleep(retryDelay)
			continue
		}

		c.ch = ch

		if withConfirm {
			err = ch.Confirm(false)
			if err != nil {
				return nil, fmt.Errorf("failed to set channel in confirm mode: %s", err)
			}
		}

		timesTried = retryTimes

		go c.keepAlive(ch.NotifyClose(make(chan *amqp091.Error)), withConfirm)

		return c, nil
	}

	return nil, fmt.Errorf("failed to create channel after %d retries: %s", retryTimes, err)
}

func (c *Channel) PublisherBuilder(exchange string, routingKey string) *PublisherBuilder {
	return newPublisherBuilder(c, c.retryTimes, c.retryDelay, exchange, routingKey)
}

func (c *Channel) ConsumerBuilder(name, queue string) *ConsumerBuilder {
	return newConsumerBuilder(c, c.retryTimes, c.retryDelay, name, queue)
}

func (c *Channel) QueueBuilder() *QueueBuilder {
	return newQueueBuilder(c, c.retryTimes, c.retryDelay)
}

func (c *Channel) ExchangeBuilder(name string) *ExchangeBuilder {
	return newExchangeBuilder(c, c.retryTimes, c.retryDelay, name, amqp091.DefaultExchange)
}

func (c *Channel) FanoutExchangeBuilder(name string) *ExchangeBuilder {
	return newExchangeBuilder(c, c.retryTimes, c.retryDelay, name, amqp091.ExchangeFanout)
}

func (c *Channel) DirectExchangeBuilder(name string) *ExchangeBuilder {
	return newExchangeBuilder(c, c.retryTimes, c.retryDelay, name, amqp091.ExchangeDirect)
}

func (c *Channel) TopicExchangeBuilder(name string) *ExchangeBuilder {
	return newExchangeBuilder(c, c.retryTimes, c.retryDelay, name, amqp091.ExchangeTopic)
}

func (c *Channel) CloseChan() <-chan error {
	return c.closeChan
}

// keepAlive keeps the channel alive
func (c *Channel) keepAlive(
	closeNotifier chan *amqp091.Error,
	withConfirm bool,
) {
	var (
		err          error
		lastCloseErr *amqp091.Error
	)

	for {
		select {
		case lastCloseErr = <-closeNotifier:
			break
		case <-c.closeChan:
			return
		}

		closeNotifier, err = c.reconnect(withConfirm)
		if err != nil {
			c.closeChan <- fmt.Errorf(
				"failed to create a channel after %d tries: %s - %s",
				c.retryTimes,
				lastCloseErr,
				err,
			)

			return
		}
	}
}

// reconnect reconnects the channel
func (c *Channel) reconnect(withConfirm bool) (chan *amqp091.Error, error) {
	c.m.Lock()
	defer c.m.Unlock()

	timesTried := c.retryTimes

	var (
		ch  *amqp091.Channel
		err error
	)

	for timesTried > 0 {
		ch, err = c.rmq.conn.Channel()
		if err != nil {
			timesTried--
			time.Sleep(c.retryDelay)
			continue
		}

		c.ch = ch

		if withConfirm {
			err = ch.Confirm(false)
			if err != nil {
				return nil, fmt.Errorf("failed to set channel in confirm mode: %s", err)
			}
		}

		timesTried = c.retryTimes

		return ch.NotifyClose(make(chan *amqp091.Error)), nil
	}

	return nil, fmt.Errorf("failed to reconnect channel after %d retries: %s", c.retryTimes, err)
}

func (c *Channel) channel() *amqp091.Channel {
	c.m.Lock()
	defer c.m.Unlock()

	return c.ch
}
