package rmq

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type ConsumerImpl interface {
	Messages() <-chan amqp091.Delivery
	ErrorChan() <-chan error
	Cancel() error
	Close() error
}

type Consumer struct {
	consumerBuilder *ConsumerBuilder

	delivery chan amqp091.Delivery
	errChan  chan error

	consumerClosedChan chan error

	// Improved lifecycle management
	ctx        context.Context
	cancel     context.CancelFunc
	closeOnce  sync.Once
	wg         sync.WaitGroup
	mu         sync.RWMutex
	isRunning  bool
	currentTag string

	// Critical fix: Track processing goroutine
	processingWg sync.WaitGroup
}

func newConsumer(
	consumerBuilder *ConsumerBuilder,
	retryCount int,
	retryDuration time.Duration,
) (*Consumer, error) {
	ctx, cancel := context.WithCancel(context.Background())

	consumer := &Consumer{
		consumerBuilder:    consumerBuilder,
		delivery:           make(chan amqp091.Delivery, 100), // Buffered to prevent blocking
		errChan:            make(chan error, 10),             // Buffered for multiple errors
		consumerClosedChan: make(chan error, 1),
		ctx:                ctx,
		cancel:             cancel,
		isRunning:          true,
	}

	consumerBuilder.channel.addCloseChannel(consumer.consumerClosedChan)

	// Start the consumer with immediate error handling
	if err := consumer.start(); err != nil {
		// If initial start fails, still return consumer but signal the error
		consumer.signalError(fmt.Errorf("initial consumer start failed: %w", err))
	}

	// Start keep-alive in a separate goroutine
	consumer.wg.Add(1)
	go consumer.keepAlive()

	return consumer, nil
}

func (c *Consumer) start() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isRunning {
		return errors.New("consumer is closed")
	}

	tried := c.consumerBuilder.retryCount
	var err error

	for tried > 0 {
		select {
		case <-c.ctx.Done():
			return c.ctx.Err()
		default:
		}

		ch := c.consumerBuilder.channel.channel()
		if ch == nil {
			tried--
			if tried > 0 {
				time.Sleep(c.consumerBuilder.retryDelay)
			}
			continue
		}

		// Set QoS if prefetch is configured
		if c.consumerBuilder.prefetch > 0 {
			err = ch.Qos(c.consumerBuilder.prefetch, 0, false)
			if err != nil {
				tried--
				if tried > 0 {
					time.Sleep(c.consumerBuilder.retryDelay)
				}
				continue
			}
		}

		// Generate unique consumer tag
		c.currentTag = fmt.Sprintf("%s-%d", c.consumerBuilder.name, time.Now().UnixNano())

		consumerChan, err := ch.Consume(
			c.consumerBuilder.queueName,
			c.currentTag,
			c.consumerBuilder.autoAck,
			c.consumerBuilder.exclusive,
			c.consumerBuilder.noLocal,
			c.consumerBuilder.noWait,
			c.consumerBuilder.args,
		)
		if err != nil {
			tried--
			if tried > 0 {
				time.Sleep(c.consumerBuilder.retryDelay)
			}
			continue
		}

		// Start processing messages
		c.processingWg.Add(1)
		go c.process(consumerChan)

		return nil
	}

	return fmt.Errorf("failed to create consumer after %d retries: %w", c.consumerBuilder.retryCount, err)
}

func (c *Consumer) Messages() <-chan amqp091.Delivery {
	return c.delivery
}

func (c *Consumer) ErrorChan() <-chan error {
	return c.errChan
}

func (c *Consumer) Cancel() error {
	c.mu.RLock()
	tag := c.currentTag
	ch := c.consumerBuilder.channel.channel()
	c.mu.RUnlock()

	if ch == nil || tag == "" {
		return errors.New("consumer not active")
	}

	return ch.Cancel(tag, false)
}

func (c *Consumer) Close() error {
	c.closeOnce.Do(func() {
		c.mu.Lock()
		c.isRunning = false
		c.mu.Unlock()

		// Cancel context first
		c.cancel()

		// Try to cancel the consumer
		c.Cancel()

		// Wait for processing goroutines to finish with timeout
		processingDone := make(chan struct{})
		go func() {
			c.processingWg.Wait()
			close(processingDone)
		}()

		// Wait for main goroutines to finish with timeout
		mainDone := make(chan struct{})
		go func() {
			c.wg.Wait()
			close(mainDone)
		}()

		// Wait with timeout
		timeout := time.NewTimer(3 * time.Second)
		defer timeout.Stop()

		select {
		case <-processingDone:
		case <-timeout.C:
			// Force close after timeout
		}

		select {
		case <-mainDone:
		case <-time.After(2 * time.Second):
			// Force close after timeout
		}

		// CRITICAL FIX: Always close delivery channel
		close(c.delivery)
		close(c.errChan)
	})
	return nil
}

func (c *Consumer) process(consumerChan <-chan amqp091.Delivery) {
	defer c.processingWg.Done()

	for {
		select {
		case <-c.ctx.Done():
			return
		case delivery, ok := <-consumerChan:
			if !ok {
				// Channel closed, signal reconnection needed
				select {
				case c.consumerClosedChan <- errors.New("consumer channel closed"):
				case <-c.ctx.Done():
				}
				return
			}

			// Try to send delivery, but don't block if consumer is closing
			select {
			case c.delivery <- delivery:
			case <-c.ctx.Done():
				return
			case <-time.After(1 * time.Second):
				// If we can't send for 1 second, consumer might be blocked
				// Log this and continue to prevent goroutine leak
				continue
			}
		}
	}
}

func (c *Consumer) keepAlive() {
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			return

		case closeErr := <-c.consumerClosedChan:
			if closeErr == nil {
				// Clean shutdown
				return
			}

			c.mu.RLock()
			running := c.isRunning
			c.mu.RUnlock()

			if !running {
				return
			}

			// Wait for current processing to finish before reconnecting
			c.processingWg.Wait()

			// Attempt to restart the consumer with exponential backoff
			backoff := time.Second
			maxBackoff := 30 * time.Second
			maxRetries := 5

			for i := 0; i < maxRetries; i++ {
				select {
				case <-c.ctx.Done():
					return
				default:
				}

				if err := c.start(); err == nil {
					// Successfully restarted
					break
				}

				if i == maxRetries-1 {
					// Final attempt failed - signal error and close delivery channel
					c.signalError(fmt.Errorf("failed to restart consumer after %d attempts: %w", maxRetries, closeErr))

					// CRITICAL: Close delivery channel to unblock range loops
					c.mu.Lock()
					if c.isRunning {
						c.isRunning = false
						close(c.delivery)
					}
					c.mu.Unlock()
					return
				}

				// Wait before retry with exponential backoff
				time.Sleep(backoff)
				backoff *= 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
			}
		}
	}
}

func (c *Consumer) signalError(err error) {
	select {
	case c.errChan <- err:
	case <-c.ctx.Done():
	case <-time.After(100 * time.Millisecond):
		// Non-blocking error send
	}
}
