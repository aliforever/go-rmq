package rmq

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

var ConnectionClosedError = fmt.Errorf("connection_closed")

type RmqImpl interface {
	SetOnError(onError func(err error))
	Connect(retryCount int, retryDelay time.Duration, onRetryError func(err error)) (<-chan error, error)
	Close() error
	NewChannel() (ChannelImpl, error)
	NewChannelWithConfirm() (ChannelImpl, error)
	IsConnected() bool
	IsHealthy() bool
}

type RMQ struct {
	address string

	connMutex sync.RWMutex
	conn      *amqp091.Connection

	retryCount int
	retryDelay time.Duration

	onRetryError func(err error)

	// Improved lifecycle management
	ctx       context.Context
	cancel    context.CancelFunc
	closeOnce sync.Once
	wg        sync.WaitGroup

	// Connection state
	connected int32 // atomic flag
	healthy   int32 // atomic flag

	// Error handling
	errChan chan error
}

func New(address string) *RMQ {
	ctx, cancel := context.WithCancel(context.Background())

	return &RMQ{
		address: address,
		ctx:     ctx,
		cancel:  cancel,
		errChan: make(chan error, 1),
	}
}

func (r *RMQ) IsConnected() bool {
	return atomic.LoadInt32(&r.connected) == 1
}

func (r *RMQ) IsHealthy() bool {
	return atomic.LoadInt32(&r.healthy) == 1
}

func (r *RMQ) SetOnError(onError func(err error)) {
	// This method can be implemented if needed for additional error handling
}

func (r *RMQ) Connect(retryCount int, retryDelay time.Duration, onRetryError func(err error)) (<-chan error, error) {
	r.connMutex.Lock()
	defer r.connMutex.Unlock()

	if onRetryError != nil {
		r.onRetryError = onRetryError
	}

	if retryCount == 0 {
		retryCount = 5
	}

	if retryDelay == 0 {
		retryDelay = 10 * time.Second
	}

	r.retryCount = retryCount
	r.retryDelay = retryDelay

	if err := r.connect(); err != nil {
		return nil, err
	}

	// Start keep-alive goroutine
	r.wg.Add(1)
	go r.keepAlive()

	return r.errChan, nil
}

func (r *RMQ) connect() error {
	tries := r.retryCount
	var err error

	for tries > 0 {
		select {
		case <-r.ctx.Done():
			return r.ctx.Err()
		default:
		}

		conn, err := amqp091.Dial(r.address)
		if err != nil {
			if r.onRetryError != nil {
				r.onRetryError(err)
			}
			tries--
			if tries > 0 {
				time.Sleep(r.retryDelay)
			}
			continue
		}

		r.conn = conn
		atomic.StoreInt32(&r.connected, 1)
		atomic.StoreInt32(&r.healthy, 1)
		return nil
	}

	atomic.StoreInt32(&r.connected, 0)
	atomic.StoreInt32(&r.healthy, 0)
	return fmt.Errorf("failed to connect to RabbitMQ after %d tries: %w", r.retryCount, err)
}

func (r *RMQ) Close() error {
	var err error
	r.closeOnce.Do(func() {
		r.cancel()

		r.connMutex.Lock()
		if r.conn != nil {
			err = r.conn.Close()
		}
		r.connMutex.Unlock()

		atomic.StoreInt32(&r.connected, 0)
		atomic.StoreInt32(&r.healthy, 0)

		// Wait for goroutines with timeout
		done := make(chan struct{})
		go func() {
			r.wg.Wait()
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(5 * time.Second):
		}

		close(r.errChan)
	})
	return err
}

func (r *RMQ) NewChannel() (*Channel, error) {
	if !r.IsConnected() {
		return nil, ConnectionClosedError
	}
	return newChannel(r, r.retryCount, r.retryDelay, false)
}

func (r *RMQ) NewChannelWithConfirm() (*Channel, error) {
	if !r.IsConnected() {
		return nil, ConnectionClosedError
	}
	return newChannel(r, r.retryCount, r.retryDelay, true)
}

func (r *RMQ) keepAlive() {
	defer r.wg.Done()

	r.connMutex.RLock()
	conn := r.conn
	r.connMutex.RUnlock()

	if conn == nil {
		return
	}

	closeNotifier := conn.NotifyClose(make(chan *amqp091.Error, 1))
	ticker := time.NewTicker(30 * time.Second) // Health check every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-r.ctx.Done():
			return

		case closeErr := <-closeNotifier:
			atomic.StoreInt32(&r.connected, 0)
			atomic.StoreInt32(&r.healthy, 0)

			if closeErr == nil {
				// Clean shutdown
				select {
				case r.errChan <- nil:
				default:
				}
				return
			}

			// Signal error
			select {
			case r.errChan <- closeErr:
			default:
			}

			// Attempt reconnection
			if err := r.reconnect(); err != nil {
				select {
				case r.errChan <- fmt.Errorf("failed to reconnect: %w", err):
				default:
				}
				return
			}

			// Set up new close notification
			r.connMutex.RLock()
			conn := r.conn
			r.connMutex.RUnlock()

			if conn == nil {
				return
			}

			closeNotifier = conn.NotifyClose(make(chan *amqp091.Error, 1))

		case <-ticker.C:
			// Periodic health check
			if !r.IsConnected() {
				if err := r.reconnect(); err != nil {
					select {
					case r.errChan <- fmt.Errorf("health check reconnection failed: %w", err):
					default:
					}
					return
				}
			}
		}
	}
}

func (r *RMQ) reconnect() error {
	backoff := time.Second
	maxBackoff := 30 * time.Second
	maxRetries := r.retryCount

	for attempt := 0; attempt < maxRetries; attempt++ {
		select {
		case <-r.ctx.Done():
			return r.ctx.Err()
		default:
		}

		r.connMutex.Lock()
		if err := r.connect(); err == nil {
			r.connMutex.Unlock()
			return nil
		}
		r.connMutex.Unlock()

		if attempt < maxRetries-1 {
			time.Sleep(backoff)
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
	}

	return fmt.Errorf("failed to reconnect after %d attempts", maxRetries)
}
