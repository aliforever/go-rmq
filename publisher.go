package rmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	genericSync "github.com/aliforever/go-generic-sync-map"
	"github.com/rabbitmq/amqp091-go"
	"sync"
	"time"
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
	ch *amqp091.Channel

	locker sync.Mutex

	exchange   string
	routingKey string

	dataType      dataType
	contentType   string
	deliveryMode  uint8
	headers       map[string]interface{}
	correlationID string
	replyToID     string
	expiration    string

	responseTimeout time.Duration

	mandatory bool
	immediate bool
}

func NewPublisher(ch *amqp091.Channel, exchange string, routingKey string) *Publisher {
	return &Publisher{ch: ch, headers: map[string]interface{}{}, exchange: exchange, routingKey: routingKey}
}

func NewPublisherWithChannel(conn *amqp091.Connection, exchange string, routingKey string) (*Publisher, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &Publisher{ch: ch, headers: map[string]interface{}{}, exchange: exchange, routingKey: routingKey}, nil
}

func (p *Publisher) SetDataTypeBytes() *Publisher {
	p.locker.Lock()
	defer p.locker.Unlock()

	p.dataType = DataTypeBytes

	return p
}

func (p *Publisher) SetDataTypeJSON() *Publisher {
	p.locker.Lock()
	defer p.locker.Unlock()

	p.dataType = DataTypeJSON
	p.contentType = "application/json"

	return p
}

func (p *Publisher) SetContentType(contentType string) *Publisher {
	p.contentType = contentType

	return p
}

func (p *Publisher) DeliveryModePersistent() *Publisher {
	p.locker.Lock()
	defer p.locker.Unlock()

	p.deliveryMode = amqp091.Persistent

	return p
}

func (p *Publisher) DeliveryModeTransient() *Publisher {
	p.locker.Lock()
	defer p.locker.Unlock()

	p.deliveryMode = amqp091.Transient

	return p
}

func (p *Publisher) AddHeader(key string, val interface{}) *Publisher {
	p.locker.Lock()
	defer p.locker.Unlock()

	p.headers[key] = val

	return p
}

func (p *Publisher) SetCorrelationID(id string) *Publisher {
	p.locker.Lock()
	defer p.locker.Unlock()

	p.correlationID = id

	return p
}

func (p *Publisher) SetReplyToID(id string) *Publisher {
	p.locker.Lock()
	defer p.locker.Unlock()

	p.replyToID = id

	return p
}

func (p *Publisher) SetExpiration(dur time.Duration) *Publisher {
	p.locker.Lock()
	defer p.locker.Unlock()

	p.expiration = fmt.Sprintf("%d", dur.Milliseconds())

	return p
}

func (p *Publisher) SetMandatory() *Publisher {
	p.locker.Lock()
	defer p.locker.Unlock()

	p.mandatory = true

	return p
}

func (p *Publisher) SetImmediate() *Publisher {
	p.locker.Lock()
	defer p.locker.Unlock()

	p.immediate = true

	return p
}

func (p *Publisher) SetResponseTimeout(timeout time.Duration) *Publisher {
	p.locker.Lock()
	defer p.locker.Unlock()

	p.responseTimeout = timeout

	return p
}

func (p *Publisher) Publish(ctx context.Context, data interface{}) error {
	p.locker.Lock()
	defer p.locker.Unlock()

	body, err := p.makeData(data)
	if err != nil {
		return err
	}

	return p.ch.PublishWithContext(ctx, p.exchange, p.routingKey, p.mandatory, p.immediate, amqp091.Publishing{
		Headers:       p.headers,
		ContentType:   p.contentType,
		DeliveryMode:  p.deliveryMode,
		CorrelationId: p.correlationID,
		ReplyTo:       p.replyToID,
		Expiration:    p.expiration,
		Body:          body,
	})
}

// PublishAwaitResponse creates a channel and stores it in responseMap
//	If an outsider writes the response to that map it checks for the reply to id to see if it matches the correlation id
//	Fails upon an invalid correlation id or in case of a timeout
func (p *Publisher) PublishAwaitResponse(
	ctx context.Context, data interface{}, responseMap *genericSync.Map[DeliveryChannel]) (amqp091.Delivery, error) {

	p.locker.Lock()
	defer p.locker.Unlock()

	if p.responseTimeout == 0 {
		return amqp091.Delivery{}, ResponseTimeoutNotSetError
	}

	if responseMap == nil {
		return amqp091.Delivery{}, ResponseMapNotSetError
	}

	if p.correlationID == "" {
		return amqp091.Delivery{}, CorrelationIdNotSetError
	}

	ch := make(chan amqp091.Delivery)

	responseMap.Store(p.correlationID, ch)

	body, err := p.makeData(data)
	if err != nil {
		return amqp091.Delivery{}, err
	}

	err = p.ch.PublishWithContext(ctx, p.exchange, p.routingKey, p.mandatory, p.immediate, amqp091.Publishing{
		Headers:       p.headers,
		ContentType:   p.contentType,
		DeliveryMode:  p.deliveryMode,
		CorrelationId: p.correlationID,
		ReplyTo:       p.replyToID,
		Expiration:    p.expiration,
		Body:          body,
	})
	if err != nil {
		return amqp091.Delivery{}, err
	}

	result, err := p.waitForResponse(ctx, p.correlationID, ch)
	if err != nil {
		return amqp091.Delivery{}, err
	}

	return result, nil
}

func (p *Publisher) PublishWithConfirmation(ctx context.Context, data interface{}) (bool, error) {
	p.locker.Lock()
	defer p.locker.Unlock()

	body, err := p.makeData(data)
	if err != nil {
		return false, err
	}

	ch, err := p.ch.PublishWithDeferredConfirmWithContext(ctx, p.exchange, p.routingKey, p.mandatory, p.immediate, amqp091.Publishing{
		Headers:       p.headers,
		ContentType:   p.contentType,
		DeliveryMode:  p.deliveryMode,
		CorrelationId: p.correlationID,
		ReplyTo:       p.replyToID,
		Expiration:    p.expiration,
		Body:          body,
	})
	if err != nil {
		return false, err
	}

	return ch.WaitContext(ctx)
}

func (p *Publisher) waitForResponse(ctx context.Context, correlationID string, ch chan amqp091.Delivery) (amqp091.Delivery, error) {
	ticker := time.NewTicker(p.responseTimeout)

	for {
		select {
		case <-ctx.Done():
			return amqp091.Delivery{}, ctx.Err()
		case <-ticker.C:
			return amqp091.Delivery{}, PublishResponseTimeoutError
		case result := <-ch:
			if result.ReplyTo != correlationID {
				return amqp091.Delivery{}, PublishResponseInvalidReplyToIdError
			}
			return result, nil
		}
	}
}

func (p *Publisher) makeData(data interface{}) (body []byte, err error) {
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
