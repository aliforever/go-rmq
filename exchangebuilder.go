package rmq

import (
	"fmt"
	"time"
)

type ExchangeBuilderImpl interface {
	DeleteOnDeclare(ifUnused, noWait bool) ExchangeBuilderImpl
	SetDurable() ExchangeBuilderImpl
	SetAutoDelete() ExchangeBuilderImpl
	SetInternal() ExchangeBuilderImpl
	SetNoWait() ExchangeBuilderImpl
	AddArg(key string, val interface{}) ExchangeBuilderImpl
	Declare() error
}

type ExchangeBuilder struct {
	channel *Channel

	exchangeType string
	name         string

	deleteOnDeclare         bool
	deleteOnDeclareIfUnused bool
	deleteOnDeclareIfNoWait bool

	retryCount int
	retryDelay time.Duration

	durable    bool
	autoDelete bool
	internal   bool
	noWait     bool
	args       map[string]interface{}
}

func newExchangeBuilder(
	channel *Channel,
	retryCount int,
	retryDuration time.Duration,
	name,
	exchangeType string,
) *ExchangeBuilder {

	return &ExchangeBuilder{
		channel:      channel,
		retryCount:   retryCount,
		retryDelay:   retryDuration,
		exchangeType: exchangeType,
		args:         map[string]interface{}{},
		name:         name,
	}
}

func (e *ExchangeBuilder) DeleteOnDeclare(ifUnused, noWait bool) *ExchangeBuilder {
	e.deleteOnDeclare = true
	e.deleteOnDeclareIfUnused = ifUnused
	e.deleteOnDeclareIfNoWait = noWait

	return e
}

func (e *ExchangeBuilder) SetDurable() *ExchangeBuilder {
	e.durable = true

	return e
}

func (e *ExchangeBuilder) SetAutoDelete() *ExchangeBuilder {
	e.autoDelete = true

	return e
}

func (e *ExchangeBuilder) SetInternal() *ExchangeBuilder {
	e.internal = true

	return e
}

func (e *ExchangeBuilder) SetNoWait() *ExchangeBuilder {
	e.noWait = true

	return e
}

func (e *ExchangeBuilder) AddArg(key string, val interface{}) *ExchangeBuilder {
	e.args[key] = val

	return e
}

func (e *ExchangeBuilder) Declare() error {
	leftTimes := e.retryCount

	var err error

	for leftTimes > 0 {
		if e.deleteOnDeclare {
			err = e.channel.channel().ExchangeDelete(e.name, e.deleteOnDeclareIfUnused, e.deleteOnDeclareIfNoWait)
			if err != nil {
				leftTimes--
				time.Sleep(e.retryDelay)
				continue
			}
		}

		if len(e.args) == 0 {
			e.args = nil
		}

		err = e.channel.channel().ExchangeDeclare(
			e.name,
			e.exchangeType,
			e.durable,
			e.autoDelete,
			e.internal,
			e.noWait,
			e.args,
		)
		if err != nil {
			leftTimes--
			time.Sleep(e.retryDelay)
			continue
		}

		return nil
	}

	return fmt.Errorf("failed to declare exchange after %d retries: %s", e.retryCount, err)
}
