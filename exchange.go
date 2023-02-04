package rmq

import "github.com/rabbitmq/amqp091-go"

type Exchange struct {
	ch *amqp091.Channel

	exchangeType string
	name         string

	deleteOnDeclare         bool
	deleteOnDeclareIfUnused bool
	deleteOnDeclareIfNoWait bool

	durable    bool
	autoDelete bool
	internal   bool
	noWait     bool
	args       map[string]interface{}
}

func NewExchange(conn *amqp091.Connection, name string) (*Exchange, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return newExchange(ch, name, amqp091.DefaultExchange), nil
}

func NewExchangeWithChannel(ch *amqp091.Channel, name string) *Exchange {
	return newExchange(ch, name, amqp091.DefaultExchange)
}

func NewFanoutExchange(conn *amqp091.Connection, name string) (*Exchange, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return newExchange(ch, name, amqp091.ExchangeFanout), nil
}

func NewFanoutExchangeWithChannel(ch *amqp091.Channel, name string) *Exchange {
	return newExchange(ch, name, amqp091.ExchangeFanout)
}

func NewDirectExchange(conn *amqp091.Connection, name string) (*Exchange, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return newExchange(ch, name, amqp091.ExchangeDirect), nil
}

func NewDirectExchangeWithChannel(ch *amqp091.Channel, name string) *Exchange {
	return newExchange(ch, name, amqp091.ExchangeDirect)
}

func NewTopicExchange(conn *amqp091.Connection, name string) (*Exchange, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return newExchange(ch, name, amqp091.ExchangeTopic), nil
}

func NewTopicExchangeWithChannel(ch *amqp091.Channel, name string) *Exchange {
	return newExchange(ch, name, amqp091.ExchangeTopic)
}

func newExchange(ch *amqp091.Channel, name, exchangeType string) *Exchange {
	return &Exchange{
		ch:           ch,
		exchangeType: exchangeType,
		args:         map[string]interface{}{},
		name:         name,
	}
}

func (e *Exchange) DeleteOnDeclare(ifUnused, noWait bool) *Exchange {
	e.deleteOnDeclare = true
	e.deleteOnDeclareIfUnused = ifUnused
	e.deleteOnDeclareIfNoWait = noWait

	return e
}

func (e *Exchange) SetDurable() *Exchange {
	e.durable = true

	return e
}

func (e *Exchange) SetAutoDelete() *Exchange {
	e.autoDelete = true

	return e
}

func (e *Exchange) SetInternal() *Exchange {
	e.internal = true

	return e
}

func (e *Exchange) SetNoWait() *Exchange {
	e.noWait = true

	return e
}

func (e *Exchange) AddArg(key string, val interface{}) *Exchange {
	e.args[key] = val

	return e
}

func (e *Exchange) Declare() error {
	if e.deleteOnDeclare {
		err := e.ch.ExchangeDelete(e.name, e.deleteOnDeclareIfUnused, e.deleteOnDeclareIfNoWait)
		if err != nil {
			return err
		}
	}

	if len(e.args) == 0 {
		e.args = nil
	}

	return e.ch.ExchangeDeclare(e.name, e.exchangeType, e.durable, e.autoDelete, e.internal, e.noWait, e.args)
}
