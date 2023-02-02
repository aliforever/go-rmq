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

func NewExchange(ch *amqp091.Channel, name string) *Exchange {
	return &Exchange{
		ch:           ch,
		exchangeType: amqp091.DefaultExchange,
		args:         map[string]interface{}{},
		name:         name,
	}
}

func NewFanoutExchange(ch *amqp091.Channel, name string) *Exchange {
	return &Exchange{
		ch:           ch,
		exchangeType: amqp091.ExchangeFanout,
		args:         map[string]interface{}{},
		name:         name,
	}
}

func NewDirectExchange(ch *amqp091.Channel, name string) *Exchange {
	return &Exchange{
		ch:           ch,
		exchangeType: amqp091.ExchangeDirect,
		args:         map[string]interface{}{},
		name:         name,
	}
}

func NewTopicExchange(ch *amqp091.Channel, name string) *Exchange {
	return &Exchange{
		ch:           ch,
		exchangeType: amqp091.ExchangeTopic,
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
