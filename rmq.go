package rmq

import (
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	"sync"
	"time"
)

var ConnectionClosedError = fmt.Errorf("connection_closed")

type RMQ struct {
	address string

	connMutex     sync.Mutex
	conn          *amqp091.Connection
	closeNotifier chan *amqp091.Error
	stopChan      chan bool

	retryCount int
	retryDelay time.Duration

	onError func(err error)
}

func New(address string) *RMQ {
	return &RMQ{
		address:       address,
		closeNotifier: make(chan *amqp091.Error),
		stopChan:      make(chan bool, 1),
	}
}

// SetOnError sets the error handler
//
// Important: It will block the reconnection so make sure to use goroutine in the callback
func (r *RMQ) SetOnError(onError func(err error)) {
	r.onError = onError
}

func (r *RMQ) Connect(retryCount int, retryDelay time.Duration) (<-chan error, error) {
	r.connMutex.Lock()
	defer r.connMutex.Unlock()

	if retryCount == 0 {
		retryCount = 5
	}

	if retryDelay == 0 {
		retryDelay = 10 * time.Second
	}

	r.retryCount = retryCount
	r.retryDelay = retryDelay

	errChan := make(chan error)

	tries := retryCount

	var (
		conn *amqp091.Connection
		err  error
	)

	for tries > 0 {
		conn, err = amqp091.Dial(r.address)
		if err != nil {
			if r.onError != nil {
				r.onError(err)
			}
			tries--
			time.Sleep(retryDelay)
			continue
		}

		r.conn = conn

		tries = retryCount

		go r.keepAlive(errChan, r.conn.NotifyClose(make(chan *amqp091.Error)), retryCount, retryDelay)

		return errChan, nil
	}

	return nil, fmt.Errorf("failed to connect to RabbitMQ after %d tries: %s", retryCount, err)
}

func (r *RMQ) Close() error {
	r.connMutex.Lock()
	defer r.connMutex.Unlock()

	defer func() {
		r.stopChan <- true
	}()

	if r.conn != nil {
		return r.conn.Close()
	}

	return nil
}

func (r *RMQ) NewChannel() (*Channel, error) {
	return newChannel(r, r.retryCount, r.retryDelay, false)
}

func (r *RMQ) NewChannelWithConfirm() (*Channel, error) {
	return newChannel(r, r.retryCount, r.retryDelay, true)
}

func (r *RMQ) keepAlive(
	errChan chan error,
	closeNotifier chan *amqp091.Error,
	retryCount int,
	retryDelay time.Duration,
) {

	timesTried := retryCount

	var (
		err          error
		lastCloseErr error
	)

	stopCh := make(chan bool)

	for {
		select {
		case <-r.stopChan:
			go func() {
				stopCh <- true
			}()
			errChan <- r.conn.Close()
			return
		case lastCloseErr = <-closeNotifier:
			go func() {
				stopCh <- false
			}()
			// wait for a second before reconnecting to make sure the connection is not forcefully closed
			time.Sleep(time.Second * 1)
			continue
		case val := <-stopCh:
			if !val {
				break
			} else {
				return
			}
		}

		closeNotifier, err = r.reconnect(timesTried, retryDelay)
		if err != nil {
			errChan <- fmt.Errorf(
				"failed to connect to RabbitMQ after %d tries: %s - %s",
				retryCount,
				lastCloseErr,
				err,
			)

			return
		}
	}
}

func (r *RMQ) reconnect(tryCount int, retryDelay time.Duration) (chan *amqp091.Error, error) {
	r.connMutex.Lock()
	defer r.connMutex.Unlock()

	timesTried := tryCount

	var lastErr error

	for timesTried > 0 {
		conn, err := amqp091.Dial(r.address)
		if err != nil {
			timesTried--
			time.Sleep(retryDelay)
			lastErr = err
			continue
		}

		r.conn = conn

		return r.conn.NotifyClose(make(chan *amqp091.Error)), nil
	}

	return nil, fmt.Errorf("failed to connect to RabbitMQ after %d tries: %s", tryCount, lastErr)
}
