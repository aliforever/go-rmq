package rmq

import (
	"errors"
	"time"
)

var (
	ConnectionNotSetError                = errors.New("connection_is_not_set")
	DataIsNotBytesError                  = errors.New("data_is_not_of_bytes_type")
	ResponseMapNotSetError               = errors.New("response_map_not_set")
	CorrelationIdNotSetError             = errors.New("correlation_id_not_set")
	PublishResponseInvalidReplyToIdError = errors.New("publish_response_invalid_reply_to_id")
)

type dataType int

const (
	DataTypeBytes dataType = iota
	DataTypeJSON
)

type PublisherBuilder struct {
	ch *Channel

	exchange   string
	routingKey string

	retryCount int
	retryDelay time.Duration

	fields *PublishFields
}

func newPublisherBuilder(
	ch *Channel,
	retryCount int,
	retryDelay time.Duration,
	exchange string,
	routingKey string,
) *PublisherBuilder {

	return &PublisherBuilder{
		ch:         ch,
		exchange:   exchange,
		routingKey: routingKey,
		fields:     NewPublishFields(),
		retryCount: retryCount,
		retryDelay: retryDelay,
	}

}

func (p *PublisherBuilder) WithFields(fields *PublishFields) *PublisherBuilder {
	p.fields = fields

	return p
}

func (p *PublisherBuilder) New() *Publisher {
	return newPublisher(p.ch, p.retryCount, p.retryDelay, p.exchange, p.routingKey, p.fields)
}

func (p *PublisherBuilder) NewWithDefaultFields() *Publisher {
	return newPublisher(p.ch, p.retryCount, p.retryDelay, p.exchange, p.routingKey, NewPublishFields())
}
