package rmq

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rabbitmq/amqp091-go"
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
	Close() error
	IsHealthy() bool
}

type Channel struct {
	mu sync.RWMutex

	rmq *RMQ
	ch  *amqp091.Channel

	retryTimes int
	retryDelay time.Duration

	closeChan     chan error
	closeChannels []chan error

	// Improved lifecycle management
	ctx       context.Context
	cancel    context.CancelFunc
	closeOnce sync.Once
	wg        sync.WaitGroup
	healthy   int32 // atomic flag for health status

	// Connection state
	isConnected int32 // atomic flag
}

func newChannel(rmq *RMQ, retryTimes int, retryDelay time.Duration, withConfirm bool) (*Channel, error) {
	ctx, cancel := context.WithCancel(context.Background())

	c := &Channel{
		rmq:        rmq,
		closeChan:  make(chan error, 1),
		retryTimes: retryTimes,
		retryDelay: retryDelay,
		ctx:        ctx,
		cancel:     cancel,
	}

	if err := c.connect(withConfirm); err != nil {
		c.Close()
		return nil, err
	}

	c.wg.Add(1)
	go c.keepAlive(withConfirm)

	return c, nil
}

func (c *Channel) connect(withConfirm bool) error {
	timesTried := c.retryTimes
	var err error

	for timesTried > 0 {
		select {
		case <-c.ctx.Done():
			return c.ctx.Err()
		default:
		}

		ch, err := c.rmq.conn.Channel()
		if err != nil {
			timesTried--
			if timesTried > 0 {
				time.Sleep(c.retryDelay)
			}
			continue
		}

		c.mu.Lock()
		c.ch = ch
		c.mu.Unlock()

		if withConfirm {
			if err = ch.Confirm(false); err != nil {
				ch.Close()
				timesTried--
				if timesTried > 0 {
					time.Sleep(c.retryDelay)
				}
				continue
			}
		}

		atomic.StoreInt32(&c.isConnected, 1)
		atomic.StoreInt32(&c.healthy, 1)
		return nil
	}

	atomic.StoreInt32(&c.isConnected, 0)
	atomic.StoreInt32(&c.healthy, 0)
	return fmt.Errorf("failed to create channel after %d retries: %w", c.retryTimes, err)
}

func (c *Channel) IsHealthy() bool {
	return atomic.LoadInt32(&c.healthy) == 1
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

func (c *Channel) Close() error {
	var err error
	c.closeOnce.Do(func() {
		c.cancel()

		c.mu.Lock()
		if c.ch != nil {
			err = c.ch.Close()
		}
		c.mu.Unlock()

		atomic.StoreInt32(&c.isConnected, 0)
		atomic.StoreInt32(&c.healthy, 0)

		// Wait for goroutines with timeout
		done := make(chan struct{})
		go func() {
			c.wg.Wait()
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(5 * time.Second):
		}

		c.signalClose(nil)
	})
	return err
}

func (c *Channel) keepAlive(withConfirm bool) {
	defer c.wg.Done()

	// Set up close notification
	c.mu.RLock()
	ch := c.ch
	c.mu.RUnlock()

	if ch == nil {
		return
	}

	closeNotifier := ch.NotifyClose(make(chan *amqp091.Error, 1))

	for {
		select {
		case <-c.ctx.Done():
			return

		case closeErr := <-closeNotifier:
			atomic.StoreInt32(&c.isConnected, 0)
			atomic.StoreInt32(&c.healthy, 0)

			if closeErr == nil {
				// Clean shutdown
				c.signalClose(nil)
				return
			}

			c.signalClose(closeErr)

			// Attempt reconnection
			if err := c.reconnect(withConfirm); err != nil {
				c.signalClose(fmt.Errorf("failed to reconnect channel: %w", err))
				return
			}

			// Set up new close notification
			c.mu.RLock()
			ch := c.ch
			c.mu.RUnlock()

			if ch == nil {
				return
			}

			closeNotifier = ch.NotifyClose(make(chan *amqp091.Error, 1))
		}
	}
}

func (c *Channel) reconnect(withConfirm bool) error {
	backoff := time.Second
	maxBackoff := 30 * time.Second
	maxRetries := c.retryTimes

	for attempt := 0; attempt < maxRetries; attempt++ {
		select {
		case <-c.ctx.Done():
			return c.ctx.Err()
		default:
		}

		if err := c.connect(withConfirm); err == nil {
			return nil
		}

		if attempt < maxRetries-1 {
			time.Sleep(backoff)
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
	}

	return fmt.Errorf("failed to reconnect channel after %d attempts", maxRetries)
}

func (c *Channel) channel() *amqp091.Channel {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ch
}

func (c *Channel) addCloseChannel(closeChan chan error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closeChannels = append(c.closeChannels, closeChan)
}

func (c *Channel) signalClose(err error) {
	c.mu.RLock()
	channels := make([]chan error, len(c.closeChannels))
	copy(channels, c.closeChannels)
	c.mu.RUnlock()

	// Signal main close channel
	select {
	case c.closeChan <- err:
	default:
	}

	// Signal all registered close channels
	for _, closeChan := range channels {
		select {
		case closeChan <- err:
		default:
		}
	}

	// Close main channel only on clean shutdown
	if err == nil {
		select {
		case <-time.After(100 * time.Millisecond):
			close(c.closeChan)
		}
	}
}
