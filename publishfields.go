package rmq

import (
	"encoding/json"
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	"time"
)

type PublishFieldsImpl interface {
	SetDataTypeBytes() PublisherBuilderImpl
	SetDataTypeJSON() PublisherBuilderImpl
	SetContentType(contentType string) PublisherBuilderImpl
	DeliveryModePersistent() PublisherBuilderImpl
	DeliveryModeTransient() PublisherBuilderImpl
	AddHeader(key string, val interface{}) PublisherBuilderImpl
	SetCorrelationID(id string) PublisherBuilderImpl
	SetReplyToID(id string) PublisherBuilderImpl
	SetExpiration(dur time.Duration) PublisherBuilderImpl
	SetMandatory() PublisherBuilderImpl
	SetImmediate() PublisherBuilderImpl
}

type PublishFields struct {
	dataType      dataType
	contentType   string
	deliveryMode  uint8
	headers       map[string]interface{}
	correlationID string
	replyToID     string
	expiration    string
	mandatory     bool
	immediate     bool
}

func NewPublishFields() *PublishFields {
	return &PublishFields{headers: map[string]interface{}{}}
}

func (p *PublishFields) SetDataTypeBytes() *PublishFields {
	p.dataType = DataTypeBytes

	return p
}

func (p *PublishFields) SetDataTypeJSON() *PublishFields {
	p.dataType = DataTypeJSON
	p.contentType = "application/json"

	return p
}

func (p *PublishFields) SetContentType(contentType string) *PublishFields {
	p.contentType = contentType

	return p
}

func (p *PublishFields) DeliveryModePersistent() *PublishFields {
	p.deliveryMode = amqp091.Persistent

	return p
}

func (p *PublishFields) DeliveryModeTransient() *PublishFields {
	p.deliveryMode = amqp091.Transient

	return p
}

func (p *PublishFields) AddHeader(key string, val interface{}) *PublishFields {
	p.headers[key] = val

	return p
}

func (p *PublishFields) SetCorrelationID(id string) *PublishFields {
	p.correlationID = id

	return p
}

func (p *PublishFields) SetReplyToID(id string) *PublishFields {
	p.replyToID = id

	return p
}

func (p *PublishFields) SetExpiration(dur time.Duration) *PublishFields {
	p.expiration = fmt.Sprintf("%d", dur.Milliseconds())

	return p
}

func (p *PublishFields) SetMandatory() *PublishFields {
	p.mandatory = true

	return p
}

func (p *PublishFields) SetImmediate() *PublishFields {
	p.immediate = true

	return p
}

func (p *PublishFields) makeData(data interface{}) (body []byte, err error) {
	var (
		ok bool
	)

	if p.dataType == DataTypeBytes {
		if body, ok = data.([]byte); !ok {
			return nil, DataIsNotBytesError
		}
	} else {
		body, err = json.Marshal(data)
		if err != nil {
			return nil, err
		}
	}

	if len(p.headers) == 0 {
		p.headers = nil
	}

	return
}
