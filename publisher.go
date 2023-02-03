package rmq

import (
	"errors"
	"github.com/rabbitmq/amqp091-go"
)

var (
	ConnectionNotSetError                = errors.New("connection_is_not_set")
	DataIsNotBytesError                  = errors.New("data_is_not_of_bytes_type")
	ResponseTimeoutNotSetError           = errors.New("response_timeout_not_set")
	ResponseMapNotSetError               = errors.New("response_map_not_set")
	CorrelationIdNotSetError             = errors.New("correlation_id_not_set")
	PublishResponseTimeoutError          = errors.New("publish_response_timed_out")
	PublishResponseInvalidReplyToIdError = errors.New("publish_response_invalid_reply_to_id")
)

type dataType int

const (
	DataTypeBytes dataType = iota
	DataTypeJSON
)

type Publisher struct {
	ch         *amqp091.Channel
	exchange   string
	routingKey string

	fields *PublishFields
}

func NewPublisherWithChannel(ch *amqp091.Channel, exchange string, routingKey string) *Publisher {
	return &Publisher{ch: ch, exchange: exchange, routingKey: routingKey, fields: NewPublishFields()}
}

func NewPublisher(conn *amqp091.Connection, exchange string, routingKey string) (*Publisher, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &Publisher{ch: ch, exchange: exchange, routingKey: routingKey, fields: NewPublishFields()}, nil
}

func (p *Publisher) WithFields(fields *PublishFields) *Publisher {
	p.fields = fields

	return p
}

func (p *Publisher) New() *Publish {
	return NewPublish(p.ch, p.exchange, p.routingKey, p.fields)
}

func (p *Publisher) NewWithDefaultFields() *Publish {
	return NewPublish(p.ch, p.exchange, p.routingKey, NewPublishFields())
}
